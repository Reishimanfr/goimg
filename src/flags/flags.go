// This file exists so that all flags are available globally and are in one place

package flags

import (
	"flag"
	"os"
	"path/filepath"
)

var (
	// Stuff related to guests and temporary files
	AllowGuestUploads = flag.Bool("allow-guest-uploads", true, "Toggles if unregistered users should be able to upload files")
	//TODO
	EnableGuestFileDeletion = flag.Bool("enable-guest-file-deletion", true, "Toggles if files uploaded by unregistered users should be deleted after set time")
	//TODO
	GuestFileDeletionTime = flag.Uint("guest-file-deletion-time", 12, "Time to wait before deleting files uploaded by unregistered users (in hours)")
	//TODO
	MaxGuestFileSize = flag.Uint("max-guest-file-size", 15, "Max file size for unregistered users (in MiB, set to 0 to make the same as registered users)")

	// Stuff related to registered users
	//TODO
	MaxFileSize = flag.Uint("max-file-size", 50, "Max file size for registered users (in MiB, set to 0 to make it unlimited [UNSUPPORTED AND NOT RECOMMENDED])")

	// General stuff related to file uploads
	// TODO
	StripMetadata = flag.Bool("strip-file-metadata", true, "Toggles if files should have their metadata stripped")
	// TODO
	FileLocation = flag.String("file-location", "on-disk", "Sets where uploaded files will be saved. Must be one of: aws, on-disk, web-dav, remote. (check wiki)")

	// User registration stuff
	// TODO
	EnableUserRegister = flag.Bool("enable-user-register", true, "Allow new users to create an account")
	// TODO
	VerifyNewUsersManually = flag.Bool("verify-new-users-manually", true, "Toggles if new users have to be verified manually before they can upload files")

	// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types/Common_types
	AllowedFileTypes = []string{
		"image/apng",
		"video/x-msvideo",
		"image/gif",
		"image/jpeg",
		"audio/mpeg",
		"video/mp4",
		"audio/ogg",
		"video/ogg",
		"image/png",
		"audio/wav",
		"audio/webm",
		"video/webm",
		"image/webp",
	}

	// JWT
	HMACSecretKey = flag.String("hmac-secret-key", "", "A random hmac secret key")

	// OVH/AWS stuff (this is only required if you set -file-location to "aws")
	UploadMaxBodySize = flag.Int("upload-max-body-size", 50, "The maximum body size for uploads (in MiB)")
	OvhEndpoint       = flag.String("ovh-endpoint", "", "Endpoint to the OVH container")
	OvhAccessToken    = flag.String("ovh-access-token", "", "Your OVH access token")
	OvhSecretKey      = flag.String("ovh-secret-key", "", "Your OVH secret key")
	OvhRegion         = flag.String("ovh-region", "", "OVH container endpoint (like 'waw' for example)")
	OvhContainerName  = flag.String("ovh-container-name", "", "Name of the container to store files in")

	// Other stuff like paths
	execPath, _ = os.Executable()
	BasePath    = filepath.Dir(execPath)
)
