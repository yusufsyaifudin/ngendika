package appsvc

import (
	"github.com/yusufsyaifudin/ngendika/internal/svc/apprepo"
	"time"
)

func AppFromRepo(app apprepo.App) App {
	a := App{
		ID:        app.ID,
		ClientID:  app.ClientID,
		Name:      app.Name,
		CreatedAt: time.UnixMicro(app.CreatedAt).UTC(),
		UpdatedAt: time.UnixMicro(app.UpdatedAt).UTC(),
		DeletedAt: time.UnixMicro(app.DeletedAt).UTC(),
	}
	return a
}
