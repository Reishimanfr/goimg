package flags

import (
	"flag"
	"os"
	"path/filepath"
)

var (
	MaxFileSize   = flag.Int64("max-file-size", 50, "Max allowed file size (in MiB). -1 means unlimited")
	StorageType   = flag.String("storage-type", "on-disk", "Sets where uploaded files get saved. Available options: on-disk, aws")
	TokenSizeBits = flag.Int("token-size-bits", 64, "Opaque token size in bits")

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

	Dev    = flag.Bool("dev", false, "Enables debugging stuff")
	Port   = flag.String("port", "8080", "Port on which the server should run. Overwritten if using secure mode")
	Secure = flag.Bool("secure", true, "Enables https")
	// Only needed if storage type is aws
	// todo

	SslCertPath = flag.String("ssl-cert-path", "", "Path to your ssl certificate file")
	SslKeyPath  = flag.String("ssl-key-path", "", "Path to your ssl certificate key file")

	execPath, _ = os.Executable()
	BasePth     = filepath.Dir(execPath)
)
