import json
import time
import pytest
import requests

sess = requests.Session()

@pytest.fixture
def create_user():
    def create_user(base_endpoint_url, email, password):
        endpoint = base_endpoint_url + "/users/register"
        resp = sess.post(endpoint, json={"email":email,"password":password})
        assert resp.status_code == 201
    return create_user

@pytest.fixture
def login():
    def login(base_endpoint_url, email, password):
        endpoint = base_endpoint_url + "/login"
        resp = sess.post(endpoint, json={"email":email,"password":password})
        assert resp.status_code == 200
        # userToken = resp.json()
        # assert userToken["token"] != ""
        # return "Bearer " + userToken["token"]
        return ""
    return login

@pytest.fixture
def create_project():
    def create_project(base_endpoint_url):
        endpoint = base_endpoint_url + "/project"
        resp = sess.post(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 201
        project = resp.json()
        assert project["id"] != ""
        assert project["status"] == "created"
        assert project["name"] != ""
        return project
    return create_project

@pytest.fixture
def get_project():
    def lol(base_endpoint_url, project_id):
        endpoint = base_endpoint_url + "/project/" + project_id
        resp = sess.get(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
        project = resp.json()
        return project
    return lol


@pytest.fixture
def update_project():
    def lol(base_endpoint_url, project_id, update_req):
        endpoint = base_endpoint_url + "/project/" + project_id
        time.sleep(2)
        resp = requests.put(endpoint, json=update_req, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
        project = resp.json()
        return project
    return lol


@pytest.fixture
def create_pdfslideimages():
    def lol(base_endpoint_url, project_id):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/pdfslideimages"
        files = {'myfile': open('tester.pdf', 'rb')}
        resp = requests.post(endpoint, files=files, cookies=sess.cookies.get_dict())
        assert resp.status_code == 201
        pdfslideimages = resp.json()
        assert pdfslideimages["id"] != ""
        assert pdfslideimages["project_id"] != ""
        assert pdfslideimages["status"] == "created"
        return pdfslideimages
    return lol


@pytest.fixture
def get_pdfslideimages():
    def lol(base_endpoint_url, project_id, pdfslideimages_id):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/pdfslideimages/" + pdfslideimages_id
        resp = requests.get(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
        pdfslidesimages = resp.json()
        assert pdfslidesimages["id"] != ""
        assert pdfslidesimages["project_id"] != ""
        return pdfslidesimages
    return lol


@pytest.fixture
def update_pdfslideimages():
    def lol(base_endpoint_url, project_id, pdfslideimages_id, update_req): 
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/pdfslideimages/" + pdfslideimages_id
        resp = requests.put(endpoint, json=update_req, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
        pdfslideimages = resp.json()
        assert pdfslideimages["id"] != ""
        assert pdfslideimages["project_id"] != ""
        return pdfslideimages
    return lol


@pytest.fixture
def create_videosegment():
    def lol(base_endpoint_url, project_id, image_id, order):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/videosegment"
        req = {
            "image_id": image_id,
            "order": order
        }
        resp = requests.post(endpoint, json=req, cookies=sess.cookies.get_dict())
        assert resp.status_code == 201
        videosegment = resp.json()
        assert videosegment["id"] != ""
        assert videosegment["project_id"] != ""
        assert videosegment["status"] == "created"
        return videosegment
    return lol


@pytest.fixture
def update_videosegment():
    def lol(base_endpoint_url, project_id, videosegment_id, req):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/videosegment/" + videosegment_id
        time.sleep(2)
        resp = requests.put(endpoint, json=req, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
        videosegment = resp.json()
        assert videosegment["id"] != ""
        assert videosegment["project_id"] != ""
        return videosegment
    return lol


@pytest.fixture
def get_videosegment():
    def lol(base_endpoint_url, project_id, videosegment_id):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/videosegment/" + videosegment_id
        resp = requests.get(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
        videosegment = resp.json()
        assert videosegment["id"] != ""
        assert videosegment["project_id"] != ""
        return videosegment
    return lol


@pytest.fixture
def videosegment_generate_video():
    def lol(base_endpoint_url, project_id, videosegment_id):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + "/videosegment/" + videosegment_id + ":generate"
        resp = requests.post(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
    return lol


@pytest.fixture
def videosegment_concat():
    def lol(base_endpoint_url, project_id):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + ":concat"
        resp = requests.post(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
    return lol


@pytest.fixture
def projectvideo_generate():
    def lol(base_endpoint_url, project_id):
        endpoint = base_endpoint_url + "/project/" + \
            project_id + ":generate-video"
        resp = requests.post(endpoint, cookies=sess.cookies.get_dict())
        assert resp.status_code == 200
    return lol


@pytest.fixture
def await_pdf_slides():
    def lol(base_endpoint_url, project_id):
        loop = 10
        current_loop = 1
        sleep_duration = 10
        endpoint = base_endpoint_url + "/project/" + project_id
        while current_loop <= loop:
            time.sleep(sleep_duration)
            resp = requests.get(endpoint, cookies=sess.cookies.get_dict())
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
    def lol(base_endpoint_url, project_id):
        loop = 10
        current_loop = 1
        sleep_duration = 10
        endpoint = base_endpoint_url + "/project/" + project_id
        while current_loop <= loop:
            time.sleep(sleep_duration)
            resp = requests.get(endpoint, cookies=sess.cookies.get_dict())
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
    def lol(base_endpoint_url, project_id):
        loop = 10
        current_loop = 1
        sleep_duration = 10
        endpoint = base_endpoint_url + "/project/" + project_id
        while current_loop <= loop:
            time.sleep(sleep_duration)
            resp = requests.get(endpoint, cookies=sess.cookies.get_dict())
            project = resp.json()
            print("Awaiting for video concat to be completed")
            if project.get("video_output_id") is not None:
                if project["video_output_id"] != "":
                    return
            current_loop = current_loop + 1
        assert False, "Awaiting for video generation to be completed"
    return lol


def test_get_project(base_endpoint, create_user, login, create_project, get_project):
    create_user(base_endpoint, "user1", "TestPassword123")
    login(base_endpoint, "user1", "TestPassword123")
    project = create_project(base_endpoint)
    project = get_project(base_endpoint, project["id"])
    assert project["id"] != ""
    assert project["status"] == "created"


def test_update_project(base_endpoint, create_user, login, create_project, get_project, update_project):
    create_user(base_endpoint, "user1-1", "TestPassword123")
    login(base_endpoint, "user1-1", "TestPassword123")
    project = create_project(base_endpoint)
    project = get_project(base_endpoint, project["id"])
    assert project["name"] != ""
    project = update_project(base_endpoint, project["id"], {"name": "new-name"})
    assert project["id"] != ""
    assert project["status"] == "created"
    assert project["name"] == "new-name"


def test_list_projects(base_endpoint, create_user, login, create_project):
    create_user(base_endpoint, "user2", "TestPassword123")
    login(base_endpoint, "user2", "TestPassword123")
    create_project(base_endpoint)
    create_project(base_endpoint)
    create_project(base_endpoint)
    endpoint = base_endpoint + "/projects"
    resp = requests.get(endpoint, cookies=sess.cookies.get_dict())
    assert resp.status_code == 200

    project_list = resp.json()
    assert len(project_list["projects"]) == 3


def test_add_pdf_slides(base_endpoint, create_user, login, create_project, create_pdfslideimages):
    create_user(base_endpoint, "user3", "TestPassword123")
    login(base_endpoint, "user3", "TestPassword123")
    project = create_project(base_endpoint)
    pdfslideimages = create_pdfslideimages(base_endpoint, project["id"])
    assert pdfslideimages["status"] == "created"
    assert pdfslideimages["id"] != ""


def test_project_on_addpdfslides(base_endpoint, create_user, login, create_project, create_pdfslideimages, get_project, await_pdf_slides):
    create_user(base_endpoint, "user4", "TestPassword123")
    login(base_endpoint, "user4", "TestPassword123")
    project = create_project(base_endpoint)
    pdfslideimages = create_pdfslideimages(base_endpoint, project["id"])
    project = get_project(base_endpoint, project["id"])
    assert project.get("pdf_slide_images") is not None
    assert len(project["pdf_slide_images"]) == 1
    project = await_pdf_slides(base_endpoint, project["id"])
    assert len(project["pdf_slide_images"][0]["slide_assets"]) == 2
    assert project["pdf_slide_images"][0]["status"] == "completed"
    assert project.get("video_segments") is not None
    assert len(project["video_segments"]) == 2


def test_project_onvideosegment(base_endpoint, create_user, login, create_project, get_project, create_videosegment):
    create_user(base_endpoint, "user5", "TestPassword123")
    login(base_endpoint, "user5", "TestPassword123")
    project = create_project(base_endpoint)
    videosegment = create_videosegment(base_endpoint, project["id"], "hahahax", 3)
    updated_project = get_project(base_endpoint, project["id"])
    assert updated_project.get("video_segments") is not None
    assert len(updated_project["video_segments"]) == 1
    assert updated_project["video_segments"][0]["id"] == videosegment["id"]
    assert updated_project["video_segments"][0]["status"] == videosegment["status"]


def test_update_script(base_endpoint, create_user, login, create_project, get_project, create_pdfslideimages, await_pdf_slides, update_videosegment):
    create_user(base_endpoint, "user6", "TestPassword123")
    login(base_endpoint, "user6", "TestPassword123")
    project = create_project(base_endpoint)
    create_pdfslideimages(base_endpoint, project["id"])
    await_pdf_slides(base_endpoint, project["id"])
    time.sleep(1)
    project = get_project(base_endpoint, project["id"])
    assert len(project["video_segments"]) == 2
    for v in project["video_segments"]:
        update_videosegment(base_endpoint, project["id"], v["id"], {"script": "hello"})
    updated_project = get_project(base_endpoint, project["id"])
    for z in updated_project["video_segments"]:
        assert z["script"] == "hello"
        assert z["status"] == "created"


def test_generate_video(
        base_endpoint, create_user, login, 
        create_project, get_project, create_pdfslideimages, 
        await_pdf_slides, update_videosegment, 
        videosegment_generate_video, await_video_generation_done):
    create_user(base_endpoint, "user7", "TestPassword123")
    login(base_endpoint, "user7", "TestPassword123")
    project = create_project(base_endpoint)
    create_pdfslideimages(base_endpoint, project["id"])
    await_pdf_slides(base_endpoint, project["id"])
    time.sleep(1)
    project = get_project(base_endpoint, project["id"])
    assert len(project["video_segments"]) == 2
    for v in project["video_segments"]:
        update_videosegment(base_endpoint, project["id"], v["id"], {"script": "hello"})
    for z in project["video_segments"]:
        videosegment_generate_video(base_endpoint, project["id"], z["id"])
    await_video_generation_done(base_endpoint, project["id"])


def test_full_flow(
        base_endpoint, create_user, login, 
        create_project, get_project,
        create_pdfslideimages, await_pdf_slides,
        update_videosegment, videosegment_generate_video, await_video_generation_done,
        videosegment_concat, await_video_concat_done):
    create_user(base_endpoint, "user8", "TestPassword123")
    login(base_endpoint, "user8", "TestPassword123")
    project = create_project(base_endpoint)
    create_pdfslideimages(base_endpoint, project["id"])
    await_pdf_slides(base_endpoint, project["id"])
    time.sleep(1)
    project = get_project(base_endpoint, project["id"])
    assert len(project["video_segments"]) == 2
    for v in project["video_segments"]:
        update_videosegment(base_endpoint, project["id"], v["id"], {
                            "script": "this is a test to check that this works"})
    for z in project["video_segments"]:
        videosegment_generate_video(base_endpoint, project["id"], z["id"])
    await_video_generation_done(base_endpoint, project["id"])
    videosegment_concat(base_endpoint, project["id"])
    await_video_concat_done(base_endpoint, project["id"])


def test_full_frontend_flow(
        base_endpoint, create_user, login, 
        create_project, get_project,
        create_pdfslideimages, await_pdf_slides,
        update_videosegment, projectvideo_generate, await_video_concat_done):
    create_user(base_endpoint, "user9", "TestPassword123")
    login(base_endpoint, "user9", "TestPassword123")
    project = create_project(base_endpoint)
    create_pdfslideimages(base_endpoint, project["id"])
    await_pdf_slides(base_endpoint, project["id"])
    time.sleep(1)
    project = get_project(base_endpoint, project["id"])
    assert len(project["video_segments"]) == 2
    for v in project["video_segments"]:
        update_videosegment(base_endpoint, project["id"], v["id"], {
                            "script": "this is a test to check that this works"})
    projectvideo_generate(base_endpoint, project["id"])
    await_video_concat_done(base_endpoint, project["id"])