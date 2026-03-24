package domain_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DementorAK/photometa/internal/domain"
)

func ExampleImageFile() {
	// This example shows how to serialize ImageFile with metadata to JSON.
	img := domain.ImageFile{
		Name: "vacation.jpg",
		Metadata: domain.Metadata{
			Format: "jpeg",
			Tags: []domain.TagInfo{
				{Type: "EXIF", Group: "Equipment", Name: "Make", Value: "Nikon"},
				{Type: "EXIF", Group: "Equipment", Name: "Model", Value: "Z6"},
			},
		},
	}

	data, err := json.MarshalIndent(img, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	os.Stdout.Write(data)

	// Output:
	// {
	//   "path": "",
	//   "name": "vacation.jpg",
	//   "metadata": {
	//     "format": "jpeg",
	//     "file_size": 0,
	//     "tags": [
	//       {
	//         "type": "EXIF",
	//         "group": "Equipment",
	//         "name": "Make",
	//         "value": "Nikon"
	//       },
	//       {
	//         "type": "EXIF",
	//         "group": "Equipment",
	//         "name": "Model",
	//         "value": "Z6"
	//       }
	//     ]
	//   }
	// }
}
