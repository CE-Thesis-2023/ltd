package reconciler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/ltd/src/internal/mqtt"
	"github.com/CE-Thesis-2023/ltd/src/service"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

type Reconciler struct {
	cameras          map[string]web.TranscoderStreamConfiguration
	cameraProperties map[string]db.Camera
	openGateConfigs  string

	controlPlaneService *service.ControlPlaneService
	deviceInfo          *configs.DeviceInfoConfigs
	commandService      *service.CommandService
	mediaService        *service.MediaController
	openGateService     *service.ProcessorController
	mqttEndpoints       *web.GetMQTTEventEndpointResponse

	mqttClient *autopaho.ConnectionManager

	updatedOpenGateConfigs string
	updatedCameras         map[string]web.TranscoderStreamConfiguration

	mu sync.Mutex
}

func NewReconciler(
	controlPlaneService *service.ControlPlaneService,
	deviceInfo *configs.DeviceInfoConfigs,
	commandService *service.CommandService,
	mediaService *service.MediaController,
	openGateService *service.ProcessorController) *Reconciler {
	if controlPlaneService == nil {
		logger.SFatal("control plane service is nil",
			zap.String("error", "control plane service is nil"))
	}
	if deviceInfo == nil {
		logger.SFatal("device info is nil",
			zap.String("error", "device info is nil"))
	}
	if commandService == nil {
		logger.SFatal("command service is nil",
			zap.String("error", "command service is nil"))
	}
	if mediaService == nil {
		logger.SFatal("media service is nil",
			zap.String("error", "media service is nil"))
	}
	if openGateService == nil {
		logger.SFatal("open gate service is nil",
			zap.String("error", "open gate service is nil"))
	}
	return &Reconciler{
		cameras:             make(map[string]web.TranscoderStreamConfiguration),
		cameraProperties:    make(map[string]db.Camera),
		controlPlaneService: controlPlaneService,
		deviceInfo:          deviceInfo,
		commandService:      commandService,
		mediaService:        mediaService,
		openGateService:     openGateService,
	}
}

func (c *Reconciler) Run(ctx context.Context) {
	logger.SDebug("reconciler loop started")

	if err := c.init(ctx); err != nil {
		logger.SError("reconciler loop initialize application failed",
			zap.Error(err))
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := c.mediaService.Reconcile(ctx); err != nil {
			logger.SFatal("media controller reconcile failed",
				zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := c.openGateService.Reconcile(ctx); err != nil {
			logger.SFatal("open gate controller reconcile failed",
				zap.Error(err))
		}
	}()

	for {
		c.mu.Lock()
		if err := c.reconcile(ctx); err != nil {
			logger.SError("reconciler loop reconcile failed",
				zap.Error(err))
			c.mu.Unlock()
			return
		}

		select {
		case <-time.After(2 * time.Second):
			c.mu.Unlock()
			continue
		case <-ctx.Done():
			logger.SInfo("reconciler loop shutdown requested")
			c.mu.Unlock()
			wg.Wait()
			return
		}
	}
}

func (c *Reconciler) init(ctx context.Context) error {
	if err := c.registerDevice(ctx); err != nil {
		logger.SError("failed to register device",
			zap.Error(err))
		return err
	}
	if err := c.pullLatestMQTTConfigurations(ctx); err != nil {
		logger.SError("failed to pull latest MQTT configurations",
			zap.Error(err))
		return err
	}
	if err := c.initializeMQTTClient(ctx); err != nil {
		logger.SError("failed to initialize MQTT client",
			zap.Error(err))
		return err
	}
	if err := c.openGateService.PrePullImages(ctx); err != nil {
		logger.SError("failed to pull images",
			zap.Error(err))
		return err
	}
	return nil
}

func (c *Reconciler) initializeMQTTClient(ctx context.Context) error {
	configs := configs.EventStoreConfigs{
		Host:       c.mqttEndpoints.Host,
		Port:       c.mqttEndpoints.Port,
		Username:   c.mqttEndpoints.Username,
		Password:   c.mqttEndpoints.Password,
		Enabled:    true,
		Level:      "info",
		TlsEnabled: c.mqttEndpoints.TlsEnabled,
		Name:       "reconciler-mqtt-client",
	}
	client := custmqtt.NewClient(
		ctx,
		custmqtt.WithClientGlobalConfigs(&configs),
		custmqtt.WithClientError(func(err error) {
			logger.SError("mqtt client error",
				zap.Error(err))
		}),
		custmqtt.WithOnReconnection(c.subscribe),
		custmqtt.WithHandlerRegister(c.registerListeners),
	)
	c.mqttClient = client
	return nil
}

func (c *Reconciler) subscribe(cm *autopaho.ConnectionManager, _ *paho.Connack) {
	subscribeOn := filepath.Join(c.mqttEndpoints.SubscribeOn, "#")
	logger.SInfo("subscribing to topic",
		zap.String("topic", subscribeOn))
	if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: subscribeOn, QoS: 1},
		},
	}); err != nil {
		logger.SFatal("failed to subscribe to topic",
			zap.Error(err))
	}
}

func (c *Reconciler) registerListeners(router *paho.StandardRouter) {
	subscribeOn := filepath.Join(c.mqttEndpoints.SubscribeOn, "#")
	logger.SInfo("registered handlers for topic",
		zap.String("topic", subscribeOn))
	router.RegisterHandler(subscribeOn, c.commandHandler)
}

func (c *Reconciler) commandHandler(p *paho.Publish) {
	logger.SDebug("received command",
		zap.String("topic", p.Topic),
		zap.String("message", string(p.Payload)))
	event := events.Event{}
	event.Parse(p.Topic)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch event.Prefix {
	case "commands":
		cameraId := event.ID
		camera, err := c.resoluteCamera(cameraId)
		if err != nil {
			logger.SError("failed to resolute camera",
				zap.Error(err))
			return
		}
		if err := c.handleCommand(ctx, &event, camera, p.Payload, p.Properties); err != nil {
			logger.SError("failed to handle command",
				zap.Error(err))
		}
	default:
		logger.SError("unknown event type",
			zap.String("event", event.Prefix))
	}
}

func (c *Reconciler) resoluteCamera(cameraId string) (*db.Camera, error) {
	camera, ok := c.cameraProperties[cameraId]
	if !ok {
		return nil, custerror.ErrorNotFound
	}
	return &camera, nil
}

func (c *Reconciler) handleCommand(ctx context.Context, event *events.Event, camera *db.Camera, payload []byte, prop *paho.PublishProperties) (err error) {
	publishTo := prop.
		ResponseTopic
	if publishTo == "" {
		logger.SDebug("response topic not found",
			zap.Reflect("event", event))
		return nil
	}
	var reply *paho.Publish

	switch event.Type {
	case "ptz":
		var req events.PTZCtrlRequest
		if err = json.Unmarshal(payload, &req); err != nil {
			return err
		}
		if err = c.commandService.PtzCtrl(ctx, camera, &req); err != nil {
			return err
		}
		reply, _ = c.buildPublish(publishTo, events.EventReply_OK, prop)

	case "info":
		var resp *hikvision.SystemDeviceInfoResponse
		resp, err = c.commandService.DeviceInfo(ctx, camera)
		if err != nil {
			return err
		}
		reply, _ = c.buildPublish(publishTo, resp, prop)
	case "healthcheck":
		var resp web.DeviceHealthcheckResponse
		reply, _ = c.buildPublish(publishTo, resp, prop)
	}
	defer func() {
		if err != nil {
			resp := events.EventReply{
				Err:    err,
				Status: err.Error(),
			}
			s, _ := resp.JSON()
			reply, _ = c.buildPublish(publishTo, s, prop)
		}
		if _, err := c.mqttClient.Publish(ctx, reply); err != nil {
			logger.SError("failed to publish response",
				zap.Error(err))
			return
		}
	}()
	return nil
}

func (c *Reconciler) buildPublish(topic string, body interface{}, receivedProperties *paho.PublishProperties) (*paho.Publish, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &paho.Publish{
		QoS:     1,
		Topic:   topic,
		Payload: payload,
		Properties: &paho.PublishProperties{
			CorrelationData: receivedProperties.CorrelationData,
			ContentType:     "application/json",
		},
	}, nil
}

func (c *Reconciler) registerDevice(ctx context.Context) error {
	err := c.controlPlaneService.
		RegisterDevice(ctx, &service.RegistrationRequest{
			DeviceId: c.
				deviceInfo.
				DeviceId,
		})
	switch err {
	case custerror.ErrorAlreadyExists:
		logger.SInfo("device already registered",
			zap.String("id", c.deviceInfo.DeviceId))
	case nil:
		logger.SInfo("device registered successfully",
			zap.String("id", c.deviceInfo.DeviceId))
	default:
		logger.SError("unable to register the device",
			zap.Error(err))
		return err
	}
	return nil
}

func (c *Reconciler) pullLatestMQTTConfigurations(ctx context.Context) error {
	mqttEndpoints, err := c.controlPlaneService.GetMQTTEndpoints(ctx, &web.GetMQTTEventEndpointRequest{
		TranscoderId: c.deviceInfo.DeviceId,
	})
	if err != nil {
		logger.SError("failed to get MQTT endpoints",
			zap.Error(err))
		return err
	}
	logger.SDebug("MQTT endpoints",
		zap.Reflect("endpoints", mqttEndpoints))

	c.mqttEndpoints = mqttEndpoints
	return nil
}

func (c *Reconciler) reconcile(ctx context.Context) error {
	logger.SDebug("reconcile Enabled")

	if err := c.pullLatestConfigurations(ctx); err != nil {
		logger.SError("failed to pull latest configurations",
			zap.Error(err))
		return err
	}

	if err := c.reconcileFFmpegStreams(); err != nil {
		logger.SError("failed to reconcile FFmpeg streams",
			zap.Error(err))
		return err
	}

	if err := c.reconcileOpenGate(); err != nil {
		logger.SError("failed to reconcile OpenGate",
			zap.Error(err))
		return err
	}

	logger.SDebug("reconcile completed")
	return nil
}

func (c *Reconciler) pullLatestConfigurations(ctx context.Context) error {
	logger.SDebug("pulling latest configurations",
		zap.String("transcoderId", c.deviceInfo.DeviceId))

	if err := c.pullOpenGateConfiguration(ctx); err != nil {
		logger.SError("failed to pull open gate configuration",
			zap.Error(err))
		return err
	}
	if err := c.pullStreamConfigurations(ctx); err != nil {
		logger.SError("failed to pull stream configurations",
			zap.Error(err))
		return err
	}
	if err := c.reconcileOpenGate(); err != nil {
		logger.SError("failed to reconcile OpenGate",
			zap.Error(err))
		return err
	}

	logger.SDebug("pulling latest configurations completed")
	return nil
}

func (c *Reconciler) pullOpenGateConfiguration(ctx context.Context) error {
	openGateResp, err := c.controlPlaneService.GetOpenGateConfigurations(ctx, &web.GetTranscoderOpenGateConfigurationRequest{
		TranscoderId: c.deviceInfo.DeviceId,
	})
	if err != nil {
		return err
	}
	decoded, err := base64.
		StdEncoding.
		DecodeString(openGateResp.Base64)
	if err != nil {
		return err
	}
	c.updatedOpenGateConfigs = string(decoded)
	return nil
}

func (c *Reconciler) pullStreamConfigurations(ctx context.Context) error {
	logger.SDebug("pulling stream configurations",
		zap.String("transcoderId", c.deviceInfo.DeviceId))
	assignedResp, err := c.controlPlaneService.GetAssignedDevices(ctx, &service.GetAssignedDevicesRequest{
		DeviceId: c.
			deviceInfo.
			DeviceId,
	})
	if err != nil {
		return err
	}
	cameras := assignedResp.Cameras

	cameraIds := make([]string, 0, len(cameras))
	for _, camera := range cameras {
		cameraIds = append(cameraIds, camera.CameraId)
		c.cameraProperties[camera.CameraId] = camera
	}

	c.updatedCameras = make(map[string]web.TranscoderStreamConfiguration)
	if len(cameraIds) > 0 {
		cameraConfigurations, err := c.controlPlaneService.GetCameraStreamSettings(ctx, &web.GetStreamConfigurationsRequest{
			CameraId: cameraIds,
		})
		if err != nil {
			return err
		}
		configs := cameraConfigurations.StreamConfigurations
		for _, config := range configs {
			c.updatedCameras[config.CameraId] = config
		}
	}
	return nil
}

func (c *Reconciler) reconcileFFmpegStreams() error {
	for cameraId, newConfig := range c.updatedCameras {
		if _, ok := c.cameras[cameraId]; !ok {
			logger.SInfo("new camera stream configuration",
				zap.String("cameraId", cameraId))
			updated, err := c.mediaService.Register(newConfig)
			if err != nil {
				logger.SError("failed to register camera stream configuration",
					zap.String("cameraId", cameraId),
					zap.Error(err))
				return err
			}
			if updated {
				logger.SInfo("camera stream configuration updated",
					zap.String("cameraId", cameraId))
			}
		}
	}
	for cameraId := range c.cameras {
		if _, ok := c.updatedCameras[cameraId]; !ok {
			logger.SInfo("camera stream configuration removed",
				zap.String("cameraId", cameraId))
			c.mediaService.Deregister(cameraId)
		}
	}
	c.cameras = c.updatedCameras
	return nil
}

func (c *Reconciler) reconcileOpenGate() error {
	if c.openGateConfigs != c.updatedOpenGateConfigs {
		logger.SInfo("OpenGate configuration updated")
		c.openGateConfigs = c.updatedOpenGateConfigs
		c.openGateService.Updates([]byte(c.openGateConfigs))
	}
	return nil
}
