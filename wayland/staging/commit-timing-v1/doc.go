package commit_timing

//go:generate go run github.com/SpinningVinyl/go-wayland/cmd/go-wayland-scanner -pkg commit_timing -prefix wp -suffix v1 -o commit_timing.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/commit-timing/commit-timing-v1.xml
