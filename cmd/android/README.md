# go-uikit android demo

This project is a small **Android demo** for **go-uikit**, built with **Ebiten Mobile** using the `apk-ebiten-builder` Makefile-based toolchain.

The goal is to let you build and run the demo on a real Android device without Android Studio.

## Prerequisites

- Go installed and available in `PATH`
- Java installed (Java 17 recommended)
- An Android device with USB debugging enabled (or [ADB over Wi-Fi](https://iqfareez.com/blog/android-wireless-debugging))

## Build & run

From the folder that contains the demo `Makefile` (the same folder as `mobile.go`):

### Install dependencies (Android SDK/NDK + ebitenmobile)

```bash
make install_dependencies
```

### Build the debug APK
```
make build
```

### Install on device and launch
```
make install
```