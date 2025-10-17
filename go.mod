module github.com/leifj/go-raid

go 1.25.1

require github.com/go-chi/chi/v5 v5.2.3

// Optional dependencies - install based on storage backend choice:
//
// For CockroachDB storage:
// require github.com/lib/pq v1.10.9
//
// For FoundationDB storage:
// require github.com/apple/foundationdb/bindings/go v0.0.0-20231216195309-3ef2e94946ee
