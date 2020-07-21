import json
import time
import pytest
import requests
from google.cloud import datastore


base_endpoint = "https://slides-to-video-manager-sj5kqt5fxq-an.a.run.app/api/v1"
project_store = "test-Project"
pdfslides_store = "test-PDFSlideImages"
videosegment_store = "test-VideoSegments"


@pytest.fixture(scope="session")
def cleanup():
    yield
    client = datastore.Client()
    query = client.query(kind=project_store)
    query_iter = query.fetch()
    # for entity in query_iter:
    #     client.delete(entity.key)


@pytest.fixture(scope="session")
def cleanup_pdfslides():
    yield
    client = datastore.Client()
    query = client.query(kind=pdfslides_store)
    query_iter = query.fetch()
    # for entity in query_iter:
    #     client.delete(entity.key)


@pytest.fixture(scope="session")
def cleanup_videosegment():
    yield
    client = datastore.Client()
    query = client.query(kind=videosegment_store)
    query_iter = query.fetch()
    # for entity in query_iter:
    #     client.delete(entity.key)


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
            data = json.loads(resp.content)
            print("Awaiting for video generation to be completed")
            print(str(current_loop) + " " + str(data))
            if data["video_output_id"] != "":
                return data
            current_loop = current_loop + 1
        assert False, "Awaiting for video generation to be completed"
    return lol


def test_get_project(create_project, get_project, cleanup):
    project = create_project
    project = get_project(project["id"])
    assert project["id"] != ""
    assert project["status"] == "created"


def test_list_projects(create_project, cleanup):
    endpoint = base_endpoint + "/projects"
    resp = requests.get(endpoint)
    assert resp.status_code == 200

    project_list = resp.json()
    assert len(project_list) == 2


# def test_update_project(create_project, update_project, get_project, cleanup):
#     project = create_project
#     updated_project = update_project(project["id"], {
#         "status": "running",
#         "idem_key": "miao"
#     })
#     assert updated_project["status"] == "running"
#     project = get_project(updated_project["id"])
#     assert project["status"] == "running"


def test_add_pdf_slides(create_project, create_pdfslideimages, cleanup):
    project = create_project
    pdfslideimages = create_pdfslideimages(project["id"])
    assert pdfslideimages["status"] == "created"
    assert pdfslideimages["id"] != ""


def test_project_on_addpdfslides(create_project, create_pdfslideimages, get_project, await_pdf_slides, cleanup):
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


# def test_update_pdf_slides(create_project, create_pdfslideimages, update_pdfslideimages, get_pdfslideimages, cleanup, cleanup_pdfslides):
#     project = create_project
#     pdfslideimages = create_pdfslideimages(project["id"])
#     zz = update_pdfslideimages(project["id"], pdfslideimages["id"], {
#         'status': 'running',
#         'idem_key': 'miao',
#     })
#     assert zz["status"] == "running"
#     reretrieve_pdfslideimages = get_pdfslideimages(
#         project["id"], pdfslideimages["id"])
#     assert reretrieve_pdfslideimages["status"] == "running"


# def test_multiple_update_pdf_slides(create_project, create_pdfslideimages, update_pdfslideimages, get_pdfslideimages, cleanup, cleanup_pdfslides):
#     project = create_project
#     pdfslideimages = create_pdfslideimages(project["id"])
#     zzz = update_pdfslideimages(project["id"], pdfslideimages["id"], {
#         "status": "running",
#         "idem_key": "miao0"
#     })
#     assert zzz["status"] == "running"
#     update_pdfslideimages(project["id"], pdfslideimages["id"], {
#         "status": "completed",
#         "idem_key": "miao1"
#     })
#     update_pdfslideimages(project["id"], pdfslideimages["id"], {
#         "status": "running",
#         "idem_key": "miao2"
#     })
#     update_pdfslideimages(project["id"], pdfslideimages["id"], {
#         "status": "completed",
#         "idem_key": "miao3"
#     })
#     reretrieve_pdfslideimages = get_pdfslideimages(
#         project["id"], pdfslideimages["id"])
#     assert reretrieve_pdfslideimages["status"] == "completed"
#     assert reretrieve_pdfslideimages["idem_key"] == "miao3"


# def test_update_videosegment(create_project, get_videosegment, create_videosegment, update_videosegment, cleanup_videosegment):
#     project = create_project
#     videosegment = create_videosegment(project["id"], "hahaha", 2)
#     req = {
#         "status": "running",
#         "idem_key": "miao"
#     }
#     updated_videosegment_z = update_videosegment(
#         project["id"], videosegment["id"], req)
#     assert updated_videosegment_z["status"] == "running"
#     req = {
#         "status": "completed",
#         "video_file": "test.mp4",
#         "idem_key": "miao2"
#     }
#     update_videosegment(project["id"], videosegment["id"], req)
#     final_videosegment = get_videosegment(project["id"], videosegment["id"])
#     assert final_videosegment["status"] == "completed"
#     assert final_videosegment["idem_key"] == "miao2"


def test_project_onvideosegment(create_project, get_project, create_videosegment, cleanup, cleanup_videosegment):
    project = create_project
    videosegment = create_videosegment(project["id"], "hahahax", 3)
    updated_project = get_project(project["id"])
    assert updated_project.get("video_segments") is not None
    assert len(updated_project["video_segments"]) == 1
    assert updated_project["video_segments"][0]["id"] == videosegment["id"]
    assert updated_project["video_segments"][0]["status"] == videosegment["status"]
