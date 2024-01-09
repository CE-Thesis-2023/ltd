package db

type Camera struct {
	CameraId string `json:"cameraId" db:"camera_id"`
	Name     string `json:"name" db:"name"`
	Ip       string `json:"ip" db:"ip"`
	Port     int    `json:"port" db:"port"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Started  bool   `json:"started" db:"started"`
}

func (c Camera) Fields() []string {
	return []string{
		"camera_id",
		"name",
		"ip",
		"port",
		"username",
		"password",
		"started",
	}
}

func (c Camera) Values() []interface{} {
	return []interface{}{
		c.CameraId,
		c.Name,
		c.Ip,
		c.Port,
		c.Username,
		c.Password,
		c.Started,
	}
}
