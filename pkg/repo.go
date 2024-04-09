package pkg

import (
    "context"
    "fmt"
    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
    "github.com/rs/zerolog/log"
    "strings"
)

type RepoOptions struct {
    KvBucket string
    ObBucket string
    Prefix   string
}

func NewRepo(nc *nats.Conn, options RepoOptions) (*Repo, error) {
    js, err := jetstream.New(nc)
    if err != nil {
        return nil, fmt.Errorf("unable to connect to jetstream: %w", err)
    }

    enc := nats.EncoderForType(nats.JSON_ENCODER)
    kv, err := js.KeyValue(context.Background(), options.KvBucket)
    if err != nil {
        return nil, fmt.Errorf("unable to get the pipline key value store: %w", err)
    }

    ob, err := js.ObjectStore(context.Background(), options.ObBucket)
    if err != nil {
        return nil, fmt.Errorf("unable to get the pipline object store: %w", err)
    }

    return &Repo{
        enc:    enc,
        kv:     kv,
        ob:     ob,
        prefix: options.Prefix,
    }, nil
}

type Repo struct {
    enc    nats.Encoder
    kv     jetstream.KeyValue
    ob     jetstream.ObjectStore
    prefix string
}

func (p *Repo) Get(ctx context.Context, key string, version string) (*PipelineVersion, error) {

    path := fmt.Sprintf("%s.pipeline.%s", p.prefix, key)
    log.Debug().Msgf("getting pipeline %s at %s", key, path)
    var ip internalPipeline
    if err := p.getKv(ctx, path, &ip); err != nil {
        return nil, fmt.Errorf("unable to get pipeline %s: %w", key, err)
    }

    versionPath := fmt.Sprintf("%s.version.%s", path, version)
    log.Debug().Msgf("getting pipeline %s version %s at %s", key, version, versionPath)
    var iv internalVersion
    if err := p.getKv(ctx, versionPath, &iv); err != nil {
        return nil, fmt.Errorf("unable to get version %s: %w", version, err)
    }

    result := &PipelineVersion{
        Key:       ip.Key,
        Name:      ip.Name,
        Version:   iv.Version,
        Artifacts: map[string][]byte{},
    }

    contentPath := fmt.Sprintf("%s/%s", strings.ReplaceAll(versionPath, ".", "/"), strings.TrimPrefix(iv.ContentKey, "/"))
    content, err := p.getOb(ctx, contentPath)
    if err != nil {
        return nil, fmt.Errorf("unable to get entrypoint %s: %w", iv.ContentKey, err)
    }
    result.Content = content

    for _, a := range iv.ArtifactKeys {
        if a == "" {
            continue
        }

        content, err := p.getOb(ctx, fmt.Sprintf("%s.%s", strings.ReplaceAll(versionPath, ".", "/"), a))
        if err != nil {
            return nil, fmt.Errorf("unable to get artifact %s: %w", a, err)
        }

        result.Artifacts[a] = content
    }

    return result, nil
}

func (p *Repo) getKv(ctx context.Context, key string, target any) error {
    e, err := p.kv.Get(ctx, key)
    if err != nil {
        return err
    }

    if err := p.enc.Decode(key, e.Value(), target); err != nil {
        return err
    }

    return nil
}

func (p *Repo) getOb(ctx context.Context, key string) ([]byte, error) {
    path := key
    if !strings.HasPrefix(path, "/") {
        path = fmt.Sprintf("/%s", path)
    }

    log.Info().Msgf("getting object %s", path)
    return p.ob.GetBytes(ctx, path)
}
