import json
import time
import pytest
import requests

base_endpoint = "http://localhost:8880/api/v1"

@pytest.fixture
def create_project():
    endpoint = base_endpoint + "/project"
    resp = requests.post(endpoint)
    assert resp.status_code == 201
    project = resp.json()
    assert project["id"] != ""
    assert project["status"] == "created"
    return project

@pytest.fixture
def get_project():
    def lol(project_id):
        endpoint = base_endpoint + "/project/" + project_id
        resp = requests.get(endpoint)
        assert resp.status_code == 200
        project = resp.json()
        return project
    return lol


@pytest.fixture
def update_project():
    def lol(project_id, update_req):
        endpoint = base_endpoint + "/project/" + project_id
        time.sleep(2)
        resp = requests.put(endpoint, json=update_req)
        assert resp.status_code == 200
        project = resp.json()
        return project
    return lol


@pytest.fixture
def create_pdfslideimages():
    def lol(project_id):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/pdfslideimages"
        files = {'myfile': open('tester.pdf', 'rb')}
        resp = requests.post(endpoint, files=files)
        assert resp.status_code == 201
        pdfslideimages = resp.json()
        assert pdfslideimages["id"] != ""
        assert pdfslideimages["project_id"] != ""
        assert pdfslideimages["status"] == "created"
        return pdfslideimages
    return lol


@pytest.fixture
def get_pdfslideimages():
    def lol(project_id, pdfslideimages_id):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/pdfslideimages/" + pdfslideimages_id
        resp = requests.get(endpoint)
        assert resp.status_code == 200
        pdfslidesimages = resp.json()
        assert pdfslidesimages["id"] != ""
        assert pdfslidesimages["project_id"] != ""
        return pdfslidesimages
    return lol


@pytest.fixture
def update_pdfslideimages():
    def lol(project_id, pdfslideimages_id, update_req):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/pdfslideimages/" + pdfslideimages_id
        resp = requests.put(endpoint, json=update_req)
        assert resp.status_code == 200
        pdfslideimages = resp.json()
        assert pdfslideimages["id"] != ""
        assert pdfslideimages["project_id"] != ""
        return pdfslideimages
    return lol


@pytest.fixture
def create_videosegment():
    def lol(project_id, image_id, order):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/videosegment"
        req = {
            "image_id": image_id,
            "order": order
        }
        resp = requests.post(endpoint, json=req)
        assert resp.status_code == 201
        videosegment = resp.json()
        assert videosegment["id"] != ""
        assert videosegment["project_id"] != ""
        assert videosegment["status"] == "created"
        return videosegment
    return lol


@pytest.fixture
def update_videosegment():
    def lol(project_id, videosegment_id, req):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/videosegment/" + videosegment_id
        time.sleep(2)
        resp = requests.put(endpoint, json=req)
        assert resp.status_code == 200
        videosegment = resp.json()
        assert videosegment["id"] != ""
        assert videosegment["project_id"] != ""
        return videosegment
    return lol


@pytest.fixture
def get_videosegment():
    def lol(project_id, videosegment_id):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/videosegment/" + videosegment_id
        resp = requests.get(endpoint)
        assert resp.status_code == 200
        videosegment = resp.json()
        assert videosegment["id"] != ""
        assert videosegment["project_id"] != ""
        return videosegment
    return lol


@pytest.fixture
def videosegment_generate_video():
    def lol(project_id, videosegment_id):
        endpoint = base_endpoint + "/project/" + \
            project_id + "/videosegment/" + videosegment_id + ":generate"
        resp = requests.post(endpoint)
        assert resp.status_code == 200
    return lol


@pytest.fixture
def videosegment_concat():
    def lol(project_id):
        endpoint = base_endpoint + "/project/" + \
            project_id + ":concat"
        resp = requests.post(endpoint)
        assert resp.status_code == 200
    return lol


@pytest.fixture
def await_pdf_slides():
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
            if data.get("pdf_slide_images") is None:
                continue
            if data["pdf_slide_images"][0]["status"] == "completed":
                return data
            current_loop = current_loop + 1
        assert False, "Awaiting for PDF Split to be ready has elapsed"
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
            project = resp.json()
            print("Awaiting for video generation to be completed")
            completed_counts = 0
            for v in project["video_segments"]:
                if v["status"] == "completed":
                    completed_counts += 1
                if completed_counts == len(project["video_segments"]):
                    return
            current_loop = current_loop + 1
        assert False, "Awaiting for video generation to be completed"
    return lol


@pytest.fixture
def await_video_concat_done():
    def lol(project_id):
        loop = 10
        current_loop = 1
        sleep_duration = 10
        endpoint = base_endpoint + "/project/" + project_id
        while current_loop <= loop:
            time.sleep(sleep_duration)
            resp = requests.get(endpoint)
            project = resp.json()
            print("Awaiting for video concat to be completed")
            if project["status"] == "completed":
                return
            current_loop = current_loop + 1
        assert False, "Awaiting for video generation to be completed"
    return lol


def test_get_project(create_project, get_project):
    project = create_project
    project = get_project(project["id"])
    assert project["id"] != ""
    assert project["status"] == "created"


def test_list_projects(create_project):
    endpoint = base_endpoint + "/projects"
    resp = requests.get(endpoint)
    assert resp.status_code == 200

    project_list = resp.json()
    assert len(project_list) == 2


def test_add_pdf_slides(create_project, create_pdfslideimages):
    project = create_project
    pdfslideimages = create_pdfslideimages(project["id"])
    assert pdfslideimages["status"] == "created"
    assert pdfslideimages["id"] != ""


def test_project_on_addpdfslides(create_project, create_pdfslideimages, get_project, await_pdf_slides):
    project = create_project
    pdfslideimages = create_pdfslideimages(project["id"])
    project = get_project(project["id"])
    assert project.get("pdf_slide_images") is not None
    assert len(project["pdf_slide_images"]) == 1
    project = await_pdf_slides(project["id"])
    assert len(project["pdf_slide_images"][0]["slide_assets"]) == 2
    assert project["pdf_slide_images"][0]["status"] == "completed"
    assert project.get("video_segments") is not None
    assert len(project["video_segments"]) == 2
