package task

import (
	"bytes"
	"encoding/csv"
	"os"
	"path"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/metadata"

	"github.com/unchartedsoftware/distil/api/util"
)

// FeaturizePrimitive will featurize the dataset fields using a primitive.
func FeaturizePrimitive(schemaFile string, index string, dataset string, config *IngestTaskConfig) error {
	sourceFolder := path.Dir(schemaFile)
	outputSchemaPath := config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative)
	outputDataPath := config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative)
	outputFolder := path.Dir(outputSchemaPath)

	// copy the source folder to have all the linked files for merging
	err := copy.Copy(sourceFolder, outputFolder)
	if err != nil {
		return errors.Wrap(err, "unable to copy source data")
	}

	// delete the existing files that will be overwritten
	os.Remove(outputSchemaPath)
	os.Remove(outputDataPath)

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	// add feature variables
	features, err := getFeatureVariables(meta, "_feature_")
	if err != nil {
		return errors.Wrap(err, "unable to get feature variables")
	}

	d3mIndexField := getD3MIndexField(mainDR)

	// open the input file
	dataPath := path.Join(sourceFolder, mainDR.ResPath)
	lines, err := readCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return errors.Wrap(err, "error reading raw data")
	}

	// add the cluster data to the raw data
	for _, f := range features {
		mainDR.Variables = append(mainDR.Variables, f.Variable)

		// header already removed, lines does not have a header
		lines, err = appendFeature(dataset, d3mIndexField, false, f, lines)
		if err != nil {
			return errors.Wrap(err, "error appending feature data")
		}
	}

	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.Name
	}
	err = writer.Write(header)
	if err != nil {
		return errors.Wrap(err, "error storing feature header")
	}

	for _, line := range lines {
		err = writer.Write(line)
		if err != nil {
			return errors.Wrap(err, "error storing feature output")
		}
	}

	// output the data with the new feature
	writer.Flush()
	err = util.WriteFileWithDirs(config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative), output.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "error writing feature output")
	}

	relativePath := getRelativePath(path.Dir(outputSchemaPath), outputDataPath)
	mainDR.ResPath = relativePath

	// write the new schema to file
	err = meta.WriteSchema(config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative))
	if err != nil {
		return errors.Wrap(err, "unable to store feature schema")
	}

	return nil
}
