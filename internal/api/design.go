package api

import (
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

type DesigntimeArtifact interface {
	Create(id string, name string, packageId string, artifactDir string) error
	Update(id string, name string, packageId string, artifactDir string) error
	Deploy(id string) error
	Delete(id string) error
	Get(id string, version string) (string, string, bool, error)
	Download(targetFile string, id string) error
	CopyContent(srcDir string, tgtDir string) error
	CompareContent(srcDir string, tgtDir string, scriptMap []string, target string) (bool, error)
}

type designtimeArtifactData struct {
	Root struct {
		Version     string `json:"Version"`
		Description string `json:"Description"`
	} `json:"d"`
}

type designtimeArtifactUpdateData struct {
	Name            string `json:"Name,omitempty"`
	Id              string `json:"Id,omitempty"`
	PackageId       string `json:"PackageId,omitempty"`
	ArtifactContent string `json:"ArtifactContent"`
}

func NewDesigntimeArtifact(artifactType string, exe *httpclnt.HTTPExecuter) DesigntimeArtifact {
	switch artifactType {
	case "MessageMapping":
		return NewMessageMapping(exe)
	case "ScriptCollection":
		return NewScriptCollection(exe)
	case "Integration":
		return NewIntegration(exe)
	case "ValueMapping":
		return NewValueMapping(exe)
	default:
		return nil
	}
}

func constructUpdateBody(method string, id string, name string, packageId string, content string) ([]byte, error) {
	artifact := &designtimeArtifactUpdateData{
		Name:            name,
		Id:              id,
		PackageId:       packageId,
		ArtifactContent: content,
	}
	// Update of Message Mapping fails as PackageId and Id are not allowed
	if method == "PUT" {
		// When updating, clear name so that it picks it up from Bundle manifest
		artifact.Name = ""
		artifact.Id = ""
		artifact.PackageId = ""
	}
	requestBody, err := json.Marshal(artifact)
	if err != nil {
		return nil, err
	}

	return requestBody, nil
}

func download(targetFile string, id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Getting content of artifact %v from tenant for comparison", id)
	content, err := getContent(id, "active", artifactType, exe)
	if err != nil {
		return err
	}

	// Create directory for target file if it doesn't exist yet
	err = os.MkdirAll(filepath.Dir(targetFile), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(targetFile, content, os.ModePerm)
	if err != nil {
		return err
	}
	log.Info().Msgf("Content of artifact %v downloaded to %v", id, targetFile)
	return nil
}

func create(id string, name string, packageId string, artifactDir string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Creating %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts", artifactType)
	return upsert(id, name, packageId, artifactDir, "POST", urlPath, 201, artifactType, "Create", exe)
}

func update(id string, name string, packageId string, artifactDir string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Updating %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", artifactType, id)
	return upsert(id, name, packageId, artifactDir, "PUT", urlPath, 200, artifactType, "Update", exe)
}

func deploy(id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Deploying %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/Deploy%vDesigntimeArtifact?Id='%s'&Version='active'", artifactType, id)
	return modifyingCall("POST", urlPath, nil, 202, fmt.Sprintf("Deploy %v designtime artifact", artifactType), exe)
}

func deleteCall(id string, artifactType string, exe *httpclnt.HTTPExecuter) error {
	log.Info().Msgf("Deleting %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='active')", artifactType, id)
	return modifyingCall("DELETE", urlPath, nil, 200, fmt.Sprintf("Delete %v designtime artifact", artifactType), exe)
}

func upsert(id string, name string, packageId string, artifactDir string, method string, urlPath string, successCode int, artifactType string, callType string, exe *httpclnt.HTTPExecuter) error {
	// Zip directory and encode to base64
	encoded, err := file.ZipDirToBase64(artifactDir)
	if err != nil {
		return err
	}
	// NOTE - PUT requires that the Id in the request matches the Bundle-SymbolicName in the MANIFEST.MF
	requestBody, err := constructUpdateBody(method, id, name, packageId, encoded)
	if err != nil {
		return err
	}

	return modifyingCall(method, urlPath, requestBody, successCode, fmt.Sprintf("%v %v designtime artifact", callType, artifactType), exe)
}

func get(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) (string, string, bool, error) {
	log.Info().Msgf("Getting details of %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')", artifactType, id, version)

	callType := fmt.Sprintf("Get %v designtime artifact", artifactType)
	resp, err := readOnlyCall(urlPath, callType, exe)
	if err != nil {
		if err.Error() == fmt.Sprintf("%v call failed with response code = 404", callType) {
			return "", "", false, nil
		} else {
			return "", "", false, err
		}
	}
	// Process response to extract version
	var jsonData *designtimeArtifactData
	respBody, err := exe.ReadRespBody(resp)
	if err != nil {
		return "", "", false, err
	}
	err = json.Unmarshal(respBody, &jsonData)
	if err != nil {
		log.Error().Msgf("Error unmarshalling response as JSON. Response body = %s", respBody)
		return "", "", false, errors.Wrap(err, 0)
	}
	return jsonData.Root.Version, jsonData.Root.Description, true, nil
}

func getContent(id string, version string, artifactType string, exe *httpclnt.HTTPExecuter) ([]byte, error) {
	log.Info().Msgf("Getting content of %v designtime artifact %v", artifactType, id)
	urlPath := fmt.Sprintf("/api/v1/%vDesigntimeArtifacts(Id='%v',Version='%v')/$value", artifactType, id, version)

	callType := fmt.Sprintf("Download %v designtime artifact", artifactType)
	resp, err := readOnlyCall(urlPath, callType, exe)
	if err != nil {
		return nil, err
	}
	return exe.ReadRespBody(resp)
}

func diffContent(firstDir string, secondDir string) bool {
	log.Info().Msg("Checking for changes in META-INF directory")
	metaDiffer := file.DiffDirectories(firstDir+"/META-INF", secondDir+"/META-INF")
	log.Info().Msg("Checking for changes in src/main/resources directory")
	resourcesDiffer := file.DiffDirectories(firstDir+"/src/main/resources", secondDir+"/src/main/resources")
	log.Info().Msg("Checking for changes in metainfo.prop")
	metainfoDiffer := DiffOptionalFile(firstDir, secondDir, "metainfo.prop")

	return metaDiffer || resourcesDiffer || metainfoDiffer
}

func copyContent(srcDir string, tgtDir string) error {
	// Copy META-INF and /src/main/resources separately so that other directories like QA, STG, PRD not copied
	err := file.ReplaceDir(srcDir+"/META-INF", tgtDir+"/META-INF")
	if err != nil {
		return err
	}
	err = file.ReplaceDir(srcDir+"/src/main/resources", tgtDir+"/src/main/resources")
	if err != nil {
		return err
	}
	// Copy also metainfo.prop that contains the description if it is available
	if file.Exists(srcDir + "/metainfo.prop") {
		err = file.CopyFile(srcDir+"/metainfo.prop", tgtDir+"/metainfo.prop")
		if err != nil {
			return err
		}
	}
	return nil
}
func DiffOptionalFile(srcDir string, tgtDir string, fileRelativePath string) bool {
	downloadedFile := fmt.Sprintf("%v/%v", srcDir, fileRelativePath)
	gitFile := fmt.Sprintf("%v/%v", tgtDir, fileRelativePath)
	if file.Exists(downloadedFile) && file.Exists(gitFile) {
		return file.DiffFile(downloadedFile, gitFile)
	} else if !file.Exists(downloadedFile) && !file.Exists(gitFile) {
		log.Warn().Msgf("Skipping diff of %v as it does not exist in both source and target", fileRelativePath)
		return false
	}
	log.Info().Msgf("File %v does not exist in either source or target", fileRelativePath)
	return true
}
