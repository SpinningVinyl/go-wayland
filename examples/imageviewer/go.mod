module github.com/SpinningVinyl/go-wayland/examples/imageviewer

go 1.19

require (
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/SpinningVinyl/go-wayland/wayland v0.0.0
	golang.org/x/image v0.3.0
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f
)

replace github.com/SpinningVinyl/go-wayland/wayland => ../../wayland
