package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
)

func (c *Client) GetProjects() ([]project.Project, error) {
	endpoint := c.mgrURL + "/api/v1/projects"
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return []project.Project{}, fmt.Errorf("Unable to retrieve projects from api endpoint")
	}
	rawProjects, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []project.Project{}, fmt.Errorf("Unable to retrieve read data")
	}
	respProjects := []project.Project{}
	err = json.Unmarshal(rawProjects, &respProjects)
	if err != nil {
		return []project.Project{}, fmt.Errorf("Unable to marshall response")
	}
	return respProjects, nil
}

func (c *Client) GetProject() (project.Project, error) {
	endpoint := c.mgrURL + "/api/v1/projects"
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return project.Project{}, fmt.Errorf("Unable to retrieve projects from api endpoint")
	}
	rawProject, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return project.Project{}, fmt.Errorf("Unable to retrieve read data")
	}
	respProject := project.Project{}
	err = json.Unmarshal(rawProject, &respProject)
	if err != nil {
		return project.Project{}, fmt.Errorf("Unable to marshall response")
	}
	return respProject, nil
}
