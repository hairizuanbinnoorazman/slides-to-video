import json
import time
import pytest
import requests


base_endpoint = "https://slides-to-video-manager-sj5kqt5fxq-an.a.run.app/api/v1"


@pytest.fixture
def create_project():
    files = {'myfile': open('tester.pdf', 'rb')}
    endpoint = base_endpoint + "/project"
    resp = requests.post(endpoint, files=files)
    return resp


@pytest.fixture
def await_project_ready():
    def lol(project_id):
        loop = 10
        current_loop = 1
        sleep_duration = 10
        endpoint = base_endpoint + "/project/" + project_id
        while current_loop <= loop:
            time.sleep(sleep_duration)
            resp = requests.get(endpoint)
            data = json.loads(resp.content)
            print("Awaiting for PDF Split to be done")
            print(str(current_loop) + " " + str(data))
            if data["slide_assets"] is not None:
                return data
            current_loop = current_loop + 1
        assert False, "Awaiting for project to be ready has elapsed"
    return lol


@pytest.fixture
def await_video_generation_done():
    def lol(project_id):
        loop = 10
        current_loop = 1
        sleep_duration = 10
        endpoint = base_endpoint + "/project/" + project_id
        while current_loop <= loop:
            time.sleep(sleep_duration)
            resp = requests.get(endpoint)
            data = json.loads(resp.content)
            print("Awaiting for video generation to be completed")
            print(str(current_loop) + " " + str(data))
            if data["video_output_id"] != "":
                return data
            current_loop = current_loop + 1
        assert False, "Awaiting for video generation to be completed"
    return lol


def test_create_project(create_project):
    resp = create_project
    assert resp.status_code == 201, "Expected response code is supposed to be 201 as created response"
    print(resp.content)


def test_get_project(create_project):
    create_proj_resp = create_project
    data = json.loads(create_proj_resp.content)

    endpoint = base_endpoint + "/project/" + data["id"]
    resp = requests.get(endpoint)
    assert resp.status_code == 200


def test_update_project(create_project, await_project_ready):
    create_proj_resp = create_project
    data = json.loads(create_proj_resp.content)

    endpoint = base_endpoint + "/project/" + data["id"]
    project_data = await_project_ready(data["id"])

    for item in project_data["slide_assets"]:
        item["text"] = "this is a test"

    update_resp = requests.put(endpoint, json=project_data)
    assert update_resp.status_code == 200, "Expected Updated response status code as 200 but its not"
    print(update_resp.content)


def test_generate_video(create_project, await_project_ready, await_video_generation_done):
    create_project_resp = create_project
    data = json.loads(create_project_resp.content)

    endpoint = base_endpoint + "/project/" + data["id"] + ":generate"
    await_project_ready(data["id"])

    generate_resp = requests.post(endpoint)
    assert generate_resp.status_code == 200, "Expected generated response status code as 200 but its not"

    await_video_generation_done(data["id"])
