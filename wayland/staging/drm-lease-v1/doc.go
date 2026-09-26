package drm_lease

//go:generate go run github.com/SpinningVinyl/go-wayland/cmd/go-wayland-scanner -pkg drm_lease -prefix wp -suffix v1 -o drm_lease.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/drm-lease/drm-lease-v1.xml
