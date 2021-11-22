package fcmrepo

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type PostgresConfig struct {
	Connection sqlx.QueryerContext `validate:"required"`
}

type Postgres struct {
	RepoFCMServerKey         RepoFCMServerKey
	RepoFCMServiceAccountKey RepoFCMServiceAccountKey
}

var _ Repo = (*Postgres)(nil)

func NewPostgres(conf PostgresConfig) (*Postgres, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	serverKey, err := NewPostgresFCMServerKey(PostgresFCMServerKeyConfig{Connection: conf.Connection})
	if err != nil {
		err = fmt.Errorf("error prepare fcm server key repo: %w", err)
		return nil, err
	}

	serviceAccKey, err := NewPostgresFCMServiceAccountKey(PostgresFCMServiceAccountKeyConfig{Connection: conf.Connection})
	if err != nil {
		err = fmt.Errorf("error prepare fcm service account key repo: %w", err)
		return nil, err
	}

	return &Postgres{
		RepoFCMServerKey:         serverKey,
		RepoFCMServiceAccountKey: serviceAccKey,
	}, nil
}

func (p *Postgres) FCMServerKey() RepoFCMServerKey {
	return p.RepoFCMServerKey
}

func (p *Postgres) FCMServiceAccountKey() RepoFCMServiceAccountKey {
	return p.RepoFCMServiceAccountKey
}
