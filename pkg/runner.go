package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
)

func Run() error {
	pipelineId := viper.GetString("pipeline")
	if pipelineId == "" {
		return fmt.Errorf("pipeline is required")
	}

	version := viper.GetString("version")
	if version == "" {
		return fmt.Errorf("version is required")
	}

	nc, err := connectNats(pipelineId, viper.GetViper())
	if err != nil {
		return fmt.Errorf("unable to connect to nats: %w", err)
	}

	repo, err := newRepo(nc, viper.GetViper())
	if err != nil {
		return fmt.Errorf("unable to create repo: %w", err)
	}

	ctx := context.Background()
	pv, err := repo.Get(ctx, pipelineId, version)
	if err != nil {
		return fmt.Errorf("unable to get pipeline: %w", err)
	}

	if err := pv.Load("/tmp"); err != nil {
		return fmt.Errorf("unable to load pipeline: %w", err)
	}

	scfg := pv.ServiceConfig()
	svc, err := micro.AddService(nc, scfg)
	if err != nil {
		return fmt.Errorf("unable to add service: %w", err)
	}
	defer svc.Stop()

	go func() {
		if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			info := svc.Info()
			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(b)
		})); err != nil {
			log.Error().Err(err).Msg("failed to start http server")
		}
	}()

	if err := pv.Run(ctx); err != nil {
		return fmt.Errorf("unable to run pipeline: %w", err)
	}

	return nil
}

func connectNats(id string, v *viper.Viper) (*nats.Conn, error) {
	repoUrl := v.GetString("repo-url")
	repoSeed := v.GetString("repo-seed")
	repoJwt := v.GetString("repo-jwt")
	repoCredsFile := v.GetString("repo-creds-file")

	if repoUrl == "" {
		return nil, fmt.Errorf("repo-url is required")
	}

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("nr-%s", id)),
	}
	if repoCredsFile != "" {
		opts = append(opts, nats.UserCredentials(repoCredsFile))
	} else if repoJwt != "" && repoSeed != "" {
		opts = append(opts, nats.UserJWTAndSeed(repoJwt, repoSeed))
	}

	return nats.Connect(repoUrl, opts...)
}

func newRepo(nc *nats.Conn, v *viper.Viper) (*Repo, error) {
	opts := RepoOptions{
		KvBucket: v.GetString("repo-kv"),
		ObBucket: v.GetString("repo-ob"),
	}

	return NewRepo(nc, opts)
}
