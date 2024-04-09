package pkg

type internalPipeline struct {
    Key  string
    Name string
}

type internalVersion struct {
    Version      string
    ContentKey   string
    ArtifactKeys []string
    Status       string
}
