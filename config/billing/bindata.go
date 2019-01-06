package billing

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assetsda35f53e42a2c065d1370dd5e00e6dfd1a288c63 = "- AmazonEC2\n- AmazonRoute53\n- AmazonECR\n- AWSLambda\n- AmazonS3\n- AmazonCloudFront\n- AmazonLightSail\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"billing"}, "/billing": []string{"servicename.yml"}}, map[string]*assets.File{
	"/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1546791225, 1546791225000000000),
		Data:     nil,
	}, "/billing": &assets.File{
		Path:     "/billing",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1546788996, 1546788996000000000),
		Data:     nil,
	}, "/billing/servicename.yml": &assets.File{
		Path:     "/billing/servicename.yml",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1546791588, 1546791588000000000),
		Data:     []byte(_Assetsda35f53e42a2c065d1370dd5e00e6dfd1a288c63),
	}}, "")
