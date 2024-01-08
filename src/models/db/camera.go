package db

type Camera struct {
	Id        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Ip        string `json:"ip" db:"ip"`
	Port      int    `json:"port" db:"port"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"password" db:"password"`
	DateAdded string `json:"date_added" db:"date_added"`
}

func (c Camera) Fields() []string {
	return []string{
		"id",
		"name",
		"ip",
		"port",
		"username",
		"password",
		"date_added",
	}
}

func (c Camera) Values() []interface{} {
	return []interface{}{
		c.Id,
		c.Name,
		c.Ip,
		c.Port,
		c.Username,
		c.Password,
		c.DateAdded,
	}
}
