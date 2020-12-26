from pytest import fixture


def pytest_addoption(parser):
    parser.addoption(
        "--base-endpoint",
        action="store"
    )


@fixture()
def base_endpoint(request):
    if request.config.getoption("--base-endpoint") is None:
        return "http://localhost:8880/api/v1"
    else:    
        return request.config.getoption("--base-endpoint")
