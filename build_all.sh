#!/bin/bash
# https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63

package=$1
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
package_split=(${package//\// })
package_name=${package_split[-1]}
if [ $package_name == "." ]; then # current directory
    package_name="pkg"
fi

platforms=("windows/amd64" "windows/386" "darwin/amd64" "linux/amd64" "linux/arm64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    echo "building $package_name for $GOOS/$GOARCH to $output_name"

    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package  #GOARM=6 ldflags '-w -extldflags "-static"'
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done