import re
from playwright.sync_api import Page, expect
import requests


TEST_EMAIL = "playwright-test@example.com"
TEST_PASSWORD = "TestPassword123"


def _register_and_login(page: Page, frontend_url: str):
    """Register a test user via the frontend-proxied API, then login via the UI."""
    requests.post(
        f"{frontend_url}/api/v1/users/register",
        json={"email": TEST_EMAIL, "password": TEST_PASSWORD},
    )
    page.goto(f"{frontend_url}/login")
    page.fill("#useremail", TEST_EMAIL)
    page.fill("#userpassword", TEST_PASSWORD)
    page.get_by_role("button", name="Login").click()
    # After login the Elm app navigates to "/" and shows "Logout" in the navbar
    expect(page.get_by_role("link", name="Logout")).to_be_visible(timeout=10000)


class TestLoginPage:
    def test_login_heading_visible(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/login")
        expect(page.get_by_role("heading", name="Login")).to_be_visible()

    def test_login_form_inputs(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/login")
        expect(page.locator("#useremail")).to_be_visible()
        expect(page.locator("#userpassword")).to_be_visible()

    def test_login_button(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/login")
        expect(page.get_by_role("button", name="Login")).to_be_visible()

    def test_register_link(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/login")
        expect(page.get_by_role("link", name="Register with Email")).to_be_visible()

    def test_google_login_link(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/login")
        expect(page.get_by_role("link", name="Google Login")).to_be_visible()


class TestRegisterPage:
    def test_register_heading(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/register")
        expect(page.get_by_role("heading", name="Register New Account")).to_be_visible()

    def test_register_form_inputs(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/register")
        expect(page.locator("#useremail")).to_be_visible()
        expect(page.locator("#userpassword")).to_be_visible()
        expect(page.locator("#confirmuserpassword")).to_be_visible()

    def test_password_mismatch_message(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/register")
        page.fill("#userpassword", "abc")
        page.fill("#confirmuserpassword", "xyz")
        expect(page.get_by_text("Passwords do not match!")).to_be_visible()

    def test_password_match_shows_register_button(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/register")
        page.fill("#userpassword", "same")
        page.fill("#confirmuserpassword", "same")
        expect(page.get_by_role("button", name="Register Account")).to_be_visible()


class TestProjectsPage:
    def test_projects_heading(self, page: Page, frontend_url: str):
        _register_and_login(page, frontend_url)
        # Use SPA navigation (click navbar link) instead of page.goto to preserve auth state
        page.get_by_role("link", name="Projects").click()
        expect(page.get_by_role("heading", name="Projects")).to_be_visible()

    def test_create_project_button(self, page: Page, frontend_url: str):
        _register_and_login(page, frontend_url)
        page.get_by_role("link", name="Projects").click()
        expect(page.get_by_role("button", name="Create Project")).to_be_visible()

    def test_projects_table_headers(self, page: Page, frontend_url: str):
        _register_and_login(page, frontend_url)
        page.get_by_role("link", name="Projects").click()
        # Create a project so the table renders
        page.get_by_role("button", name="Create Project").click()
        page.wait_for_url(re.compile(r".*/projects/.*"))
        # Navigate back to projects list via navbar
        page.get_by_role("link", name="Projects").click()
        for header_text in ["Name", "Date Created", "Last Modified"]:
            expect(page.get_by_role("columnheader", name=header_text)).to_be_visible()


class TestSingleProjectPage:
    def _create_project_and_navigate(self, page: Page, frontend_url: str):
        """Helper: login, create project via UI, navigate to its page."""
        _register_and_login(page, frontend_url)
        page.get_by_role("link", name="Projects").click()
        page.get_by_role("button", name="Create Project").click()
        page.wait_for_url(re.compile(r".*/projects/.*"))

    def test_project_heading(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.get_by_role("heading", name="Project")).to_be_visible()

    def test_project_name_input(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.locator("#projectname")).to_be_visible()

    def test_rename_button(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.get_by_role("button", name="Rename project")).to_be_visible()

    def test_upload_pdf_heading(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.get_by_role("heading", name="Step 1: Upload PDF")).to_be_visible()

    def test_upload_pdf_button(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.get_by_role("button", name="Upload PDF Slides")).to_be_visible()

    def test_generate_video_heading(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.get_by_role("heading", name="Step 2: Generate Video")).to_be_visible()

    def test_generate_video_button(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.get_by_role("button", name="Generate Video")).to_be_visible()

    def test_file_input_present(self, page: Page, frontend_url: str):
        self._create_project_and_navigate(page, frontend_url)
        expect(page.locator("input[type='file']")).to_be_visible()


class TestNavbar:
    def test_navbar_links(self, page: Page, frontend_url: str):
        _register_and_login(page, frontend_url)
        expect(page.get_by_role("link", name="Dashboard")).to_be_visible()
        expect(page.get_by_role("link", name="Projects")).to_be_visible()

    def test_logout_link_when_logged_in(self, page: Page, frontend_url: str):
        _register_and_login(page, frontend_url)
        expect(page.get_by_role("link", name="Logout")).to_be_visible()

    def test_login_link_when_logged_out(self, page: Page, frontend_url: str):
        page.goto(f"{frontend_url}/login")
        expect(page.get_by_role("link", name="Login", exact=True)).to_be_visible()
