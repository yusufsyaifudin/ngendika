package httptyped

import (
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
	"time"
)

type AppEntity struct {
	ID        int64     `json:"id"`
	ClientID  string    `json:"client_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func AppEntityFromSvc(app appsvc.App) AppEntity {
	return AppEntity{
		ID:        app.ID,
		ClientID:  app.ClientID,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
	}
}
