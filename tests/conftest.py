from pytest import fixture


def pytest_addoption(parser):
    parser.addoption(
        "--base-endpoint",
        action="store"
    )
    parser.addoption(
        "--frontend-url",
        action="store",
        default="http://localhost:8081",
        help="Base URL for the frontend application"
    )


@fixture()
def base_endpoint(request):
    if request.config.getoption("--base-endpoint") is None:
        return "http://localhost:8880/api/v1"
    else:
        return request.config.getoption("--base-endpoint")


@fixture()
def frontend_url(request):
    return request.config.getoption("--frontend-url")
