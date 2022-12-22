package mailclient

type EmailCredential struct {
	Protocol     string `json:"protocol" validate:"required,oneof=smtp"` // smtp, ...
	ServerHost   string `json:"server_host" validate:"required"`
	ServerPort   int    `json:"server_port" validate:"required"`
	AuthIdentity string `json:"auth_identity" validate:"-"` //  Authorization identity may be left blank to indicate that it is the same as the username.
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
}
