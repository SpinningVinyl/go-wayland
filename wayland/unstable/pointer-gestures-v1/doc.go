package pointer_gestures

//go:generate go run github.com/SpinningVinyl/go-wayland/cmd/go-wayland-scanner -pkg pointer_gestures -prefix zwp -suffix v1 -o pointer_gestures.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/unstable/pointer-gestures/pointer-gestures-unstable-v1.xml
