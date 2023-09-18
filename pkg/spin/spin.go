package spin

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
)

func Load(manifestLocation string) (Manifest, error) {
	m := Manifest{}
	spinTomlContents, err := os.ReadFile(manifestLocation)
	if err != nil {
		return m, fmt.Errorf("unable to open spin.toml file: %w", err)
	}

	m, err = load(spinTomlContents)
	if err != nil {
		return m, fmt.Errorf("unable to load spin.toml: %w", err)
	}

	return m, nil
}

func load(spinTomlContents []byte) (Manifest, error) {
	m := Manifest{}

	_, err := toml.Decode(string(spinTomlContents), &m)
	if err != nil {
		return m, fmt.Errorf("failed to decode spin.toml decode: %w", err)
	}

	rm := rawManifest{}
	_, err = toml.Decode(string(spinTomlContents), &rm)
	if err != nil {
		return m, fmt.Errorf("failed to decode raw spin.toml decode: %w", err)
	}

	for i, c := range rm.Components {

		componentSource, err := extractSource(c)
		if err != nil {
			return m, fmt.Errorf("extracting source on component %d: %w", i, err)
		}
		m.Components[i].Source = componentSource

		componentFiles, err := extractFiles(c)
		if err != nil {
			return m, fmt.Errorf("extracting files on component %d: %w", i, err)
		}
		m.Components[i].Files = componentFiles

		switch v := c.Source.(type) {
		case *ComponentSourceURL:
			fmt.Printf("url: %s\n", v.Url)
		}
	}

	return m, nil
}

func extractSource(c rawComponent) (ComponentSource, error) {
	sumTypeSource := ComponentSource{}
	rawSource := c.Source
	if rawSource == nil {
		return sumTypeSource, nil
	}
	v := reflect.ValueOf(rawSource)
	switch v.Kind() {
	case reflect.String:
		sourceString, ok := rawSource.(string)
		if ok {
			sumTypeSource.StringSource = ComponentSourceString(sourceString)
			break
		}
	case reflect.Map:
		stringMap, ok := rawSource.(map[string]interface{})
		if !ok {
			return sumTypeSource, fmt.Errorf("casting component to map[string]interface{}")
		}

		url, ok := stringMap["url"].(string)
		if !ok {
			return sumTypeSource, fmt.Errorf("casting url on component")
		}
		digest, ok := stringMap["digest"].(string)
		if !ok {
			return sumTypeSource, fmt.Errorf("casting digest on component")
		}
		sumTypeSource.URLSource = ComponentSourceURL{
			Url:    url,
			Digest: digest,
		}
	default:
		return sumTypeSource, fmt.Errorf("unknown component source type: %v", v.Kind())
	}
	return sumTypeSource, nil
}

func extractFiles(c rawComponent) (ComponentFiles, error) {
	sumTypeFiles := ComponentFiles{}
	if c.Files == nil {
		return sumTypeFiles, nil
	}
	rawFiles := c.Files
	v := reflect.ValueOf(rawFiles)
	switch v.Kind() {
	case reflect.String:
		fileString, ok := rawFiles.(string)
		if ok {
			sumTypeFiles.StringFiles = append(sumTypeFiles.StringFiles, ComponentFileString(fileString))
			break
		}
	case reflect.Map:
		cfm, err := extractFileMap(rawFiles)
		if err != nil {
			return sumTypeFiles, fmt.Errorf("extracting file map on component: %w", err)
		}

		sumTypeFiles.MapFiles = append(sumTypeFiles.MapFiles, cfm)

	case reflect.Slice:

		rawSlice, ok := rawFiles.([]interface{})
		if !ok {
			return sumTypeFiles, fmt.Errorf("casting files to []interface{}")
		}
		mapFiles, stringFiles, err := extractFilesFromSlice(rawSlice)
		if err != nil {
			return sumTypeFiles, fmt.Errorf("extracting files from slice: %w", err)
		}

		sumTypeFiles.MapFiles = append(sumTypeFiles.MapFiles, mapFiles...)
		sumTypeFiles.StringFiles = append(sumTypeFiles.StringFiles, stringFiles...)

	default:
		return sumTypeFiles, fmt.Errorf("unknown component source type: %v", v.Kind())
	}
	return sumTypeFiles, nil
}

func extractFilesFromSlice(rawSlice []interface{}) ([]ComponentFileMap, []ComponentFileString, error) {
	mapFiles := []ComponentFileMap{}
	stringFiles := []ComponentFileString{}

	for _, rawFile := range rawSlice {
		v := reflect.ValueOf(rawFile)
		switch v.Kind() {
		case reflect.String:
			fileString, ok := rawFile.(string)
			if ok {
				stringFiles = append(stringFiles, ComponentFileString(fileString))
				break
			}
		case reflect.Map:
			cfm, err := extractFileMap(rawFile)
			if err != nil {
				return mapFiles, stringFiles, fmt.Errorf("extracting file map on component: %w", err)
			}

			mapFiles = append(mapFiles, cfm)
		}

	}

	return mapFiles, stringFiles, nil
}

func extractFileMap(rfm interface{}) (ComponentFileMap, error) {
	cfm := ComponentFileMap{}
	stringMap, ok := rfm.(map[string]interface{})
	if !ok {
		return cfm, fmt.Errorf("casting files to map[string]interface{}")
	}

	source, ok := stringMap["source"].(string)
	if !ok {
		return cfm, fmt.Errorf("casting source on files")
	}
	destination, ok := stringMap["destination"].(string)
	if !ok {
		return cfm, fmt.Errorf("casting destination on component")
	}
	newComponentFileMap := ComponentFileMap{
		Source:      source,
		Destination: destination,
	}
	return newComponentFileMap, nil
}
