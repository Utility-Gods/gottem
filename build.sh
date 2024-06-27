#!/bin/bash

# Set the name of your application
APP_NAME="gottem"

# Set the version number
VERSION="1.0.0"

# Set the architectures to build for
ARCHITECTURES=("amd64" "386")

# Set the operating systems to build for
OPERATING_SYSTEMS=("linux" "darwin" "windows")  # "darwin" represents macOS

# Set the path to your Go project directory
PROJECT_DIR="/Users/siddharthjain/Desktop/code/utility-gods/gottem"

# Function to build and package the application
build_and_package() {
    OS=$1
    ARCH=$2

    # Set the output binary name
    BINARY_NAME="${APP_NAME}-${OS}-${ARCH}"
    if [ "$OS" == "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    # Build the application
    echo "Building ${BINARY_NAME}..."
    GOOS=$OS GOARCH=$ARCH go build -o "build/$BINARY_NAME" "$PROJECT_DIR"

    # Compress the binary
    echo "Packaging ${BINARY_NAME}..."
    cd build
    if [ "$OS" == "windows" ]; then
        zip "${BINARY_NAME}-${VERSION}.zip" $BINARY_NAME
    else
        tar -czf "${BINARY_NAME}-${VERSION}.tar.gz" $BINARY_NAME
    fi

    # Remove the binary
    rm $BINARY_NAME
    cd ..
}

# Create a build directory
mkdir -p build

# Iterate over the operating systems and architectures and build the application
for OS in "${OPERATING_SYSTEMS[@]}"
do
    for ARCH in "${ARCHITECTURES[@]}"
    do
        build_and_package $OS $ARCH
    done
done

echo "Build and packaging completed."
