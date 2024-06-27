#!/bin/bash

set -e

# Configuration
BINARY_NAME="gottem"
CURRENT_VERSION=$(cat VERSION || echo "0.0.0")
VERSION_FILE="VERSION"
PLATFORMS=("windows/amd64" "darwin/amd64" "linux/amd64")
GITHUB_REPO="Utility-Gods/gottem"  # Replace with your GitHub username/repo

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
if [ $# -lt 1 ]; then
    echo "Usage: $0 <major|minor|patch> [changelog]"
    exit 1
fi

RELEASE_TYPE=$1
NEW_VERSION=$(increment_version $CURRENT_VERSION $RELEASE_TYPE)

# Get changelog
if [ $# -ge 2 ]; then
    CHANGELOG="${@:2}"
else
    CHANGELOG="Release version $NEW_VERSION"
fi

echo "Current version: $CURRENT_VERSION"
echo "New version: $NEW_VERSION"
echo "Changelog: $CHANGELOG"
echo -n "Do you want to proceed with the release? (y/n): "
read confirm

if [ "$confirm" != "y" ]; then
    echo "Release cancelled."
    exit 0
fi

# Update VERSION file
echo $NEW_VERSION > $VERSION_FILE

# Update version in code
sed -i '' "s/const Version = \".*\"/const Version = \"$NEW_VERSION\"/" version.go

# Commit version update
git add $VERSION_FILE version.go
git commit -m "Bump version to $NEW_VERSION

Changelog:
$CHANGELOG"
git tag -a "v$NEW_VERSION" -m "Version $NEW_VERSION

Changelog:
$CHANGELOG"

# Create temporary release directory
RELEASE_DIR=$(mktemp -d)
echo "Created temporary directory: $RELEASE_DIR"

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

$CHANGELOG

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
ZIP_FILE="gottem_v${NEW_VERSION}.zip"
(cd $RELEASE_DIR && zip -r ../$ZIP_FILE .)

echo "Release v${NEW_VERSION} created successfully!"
echo "Release archive: $ZIP_FILE"

# Clean up
rm -rf $RELEASE_DIR

# Push to remote
echo -n "Do you want to push this release to remote and create a GitHub release? (y/n): "
read push_confirm

if [ "$push_confirm" = "y" ]; then
    git push origin main
    git push origin v$NEW_VERSION

    # Create GitHub release
    echo -n "Enter your GitHub personal access token: "
    read -s GITHUB_TOKEN
    echo

    # Create release
    release_response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
         --data @- \
         "https://api.github.com/repos/$GITHUB_REPO/releases" << EOF
{
  "tag_name": "v$NEW_VERSION",
  "target_commitish": "main",
  "name": "v$NEW_VERSION",
  "body": "$CHANGELOG",
  "draft": false,
  "prerelease": false
}
EOF
)

    # Extract release ID from response using string manipulation
    release_id=$(echo "$release_response" | grep -o '"id": [0-9]*' | head -1 | awk '{print $2}')

    if [ -n "$release_id" ]; then
        echo "GitHub release created successfully."

        # Upload asset
        upload_response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
             -H "Content-Type: application/zip" \
             --data-binary @"$ZIP_FILE" \
             "https://uploads.github.com/repos/$GITHUB_REPO/releases/$release_id/assets?name=$ZIP_FILE")

        if echo "$upload_response" | grep -q '"state": "uploaded"'; then
            echo "Release asset uploaded successfully."
        else
            echo "Failed to upload release asset. Please check and try again."
        fi
    else
        echo "Failed to create GitHub release. Please check your token and try again."
        echo "API Response: $release_response"
    fi

    echo "Changes pushed to remote and GitHub release created."
else
    echo "Remember to push your changes and tags to remote."
    echo "Also, don't forget to manually create a release on GitHub and upload $ZIP_FILE."
fi
