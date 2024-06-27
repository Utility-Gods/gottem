#!/bin/bash

set -e

# Configuration
BINARY_NAME="gottem"
CURRENT_VERSION=$(cat VERSION || echo "0.0.0")
VERSION_FILE="VERSION"
PLATFORMS=("windows/amd64" "darwin/amd64" "linux/amd64")

# Function to increment version
increment_version() {
    local version=$1
    local release_type=$2

    IFS='.' read -ra ver <<< "$version"
    major=${ver[0]}
    minor=${ver[1]}
    patch=${ver[2]}

    case $release_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            echo "Invalid release type. Use major, minor, or patch."
            exit 1
            ;;
    esac

    echo "${major}.${minor}.${patch}"
}

# Check if release type is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <major|minor|patch>"
    exit 1
fi

RELEASE_TYPE=$1
NEW_VERSION=$(increment_version $CURRENT_VERSION $RELEASE_TYPE)

echo "Current version: $CURRENT_VERSION"
echo "New version: $NEW_VERSION"
echo -n "Do you want to proceed with the release? (y/n): "
read confirm

if [ "$confirm" != "y" ]; then
    echo "Release cancelled."
    exit 0
fi

# Update VERSION file
echo $NEW_VERSION > $VERSION_FILE

# Update version in code (assuming you have a version.go file)
sed -i '' "s/const Version = \".*\"/const Version = \"$NEW_VERSION\"/" version.go
# sed -i '' "s/const Version = .*/const Version = \"$NEW_VERSION\"/" version.go

# Commit version update
git add $VERSION_FILE version.go
git commit -m "Bump version to $NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Version $NEW_VERSION"

# Create release directory
RELEASE_DIR="release/gottem_v${NEW_VERSION}"
mkdir -p $RELEASE_DIR

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -ra PARTS <<< "$platform"
    GOOS=${PARTS[0]}
    GOARCH=${PARTS[1]}
    output_name=$BINARY_NAME
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    echo "Building for $GOOS $GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o $RELEASE_DIR/${output_name}_${GOOS}_${GOARCH}
done

# Copy assets and create README
cp -r assets $RELEASE_DIR/ 2>/dev/null || :
cat << EOF > $RELEASE_DIR/README.md
# Gottem v${NEW_VERSION}

This is the Gottem application version ${NEW_VERSION}.

## Running the Application

Choose the appropriate executable for your platform:
- Windows: ${BINARY_NAME}_windows_amd64.exe
- macOS: ${BINARY_NAME}_darwin_amd64
- Linux: ${BINARY_NAME}_linux_amd64

Run the executable from the command line.

For more information, visit: [Your project URL]

## Changelog

[Add your changelog here]

EOF

# Create installation scripts
cat << EOF > $RELEASE_DIR/install.sh
#!/bin/bash
mkdir -p ~/.config/gottem
cp gottem_\$(uname -s | tr '[:upper:]' '[:lower:]')_amd64 /usr/local/bin/gottem
chmod +x /usr/local/bin/gottem
echo "Gottem v${NEW_VERSION} installed successfully!"
EOF

cat << EOF > $RELEASE_DIR/install.bat
@echo off
mkdir "%USERPROFILE%\.config\gottem"
copy gottem_windows_amd64.exe "%USERPROFILE%\AppData\Local\Microsoft\WindowsApps\gottem.exe"
echo Gottem v${NEW_VERSION} installed successfully!
EOF

# Create zip archive
(cd release && zip -r gottem_v${NEW_VERSION}.zip gottem_v${NEW_VERSION})

echo "Release v${NEW_VERSION} created successfully!"
echo "Release archive: release/gottem_v${NEW_VERSION}.zip"

# Optionally push to remote
echo -n "Do you want to push this release to remote? (y/n): "
read push_confirm

if [ "$push_confirm" = "y" ]; then
    git push origin main
    git push origin v$NEW_VERSION
    echo "Changes pushed to remote."
else
    echo "Remember to push your changes and tags to remote."
fi
