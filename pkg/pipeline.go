package pkg

import (
    "context"
    "fmt"
    _ "github.com/benthosdev/benthos/v4/public/components/all"
    "github.com/benthosdev/benthos/v4/public/service"
    "github.com/nats-io/nats.go/micro"
    "github.com/rs/zerolog/log"
    "os"
    "path"
)

type PipelineVersion struct {
    Key         string
    Name        string
    Description string
    Version     string

    Content []byte

    Artifacts map[string][]byte
}

func (pv *PipelineVersion) ServiceConfig() micro.Config {
    return micro.Config{
        Name:        fmt.Sprintf("benthos-%s", pv.Key),
        Version:     pv.Version,
        Description: pv.Description,
        QueueGroup:  pv.Key,
    }
}

func (pv *PipelineVersion) Load(workDir string) error {
    log.Info().Str("pipeline", pv.Key).Str("version", pv.Version).Msgf("loading artifacts into %s", workDir)

    for k, v := range pv.Artifacts {
        log.Info().Str("pipeline", pv.Key).Str("version", pv.Version).Msgf("loading artifact %s", k)
        targetPath := path.Join(workDir, k)
        targetDir := path.Dir(targetPath)
        if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
            return fmt.Errorf("unable to create directory %s: %w", targetDir, err)
        }

        if err := os.WriteFile(targetPath, v, os.ModePerm); err != nil {
            return fmt.Errorf("unable to write file %s: %w", targetPath, err)
        }
    }

    return nil
}

func (pv *PipelineVersion) Run(ctx context.Context) error {
    env := service.GlobalEnvironment().Clone()

    sb := env.NewStreamBuilder()
    if err := sb.SetYAML(string(pv.Content)); err != nil {
        return fmt.Errorf("unable to set yaml: %w", err)
    }

    stream, err := sb.Build()
    if err != nil {
        return fmt.Errorf("unable to build stream: %w", err)
    }

    return stream.Run(ctx)
}
