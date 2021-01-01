import pytest
from google.cloud import datastore

project_store = "ProjectTable"
pdfslides_store = "PDFSlideTable"
videosegment_store = "VideoSegmentsTable"

def test_util_cleanup(cleanup, cleanup_pdfslides, cleanup_videosegment):
    assert 1 == 1

@pytest.fixture(scope="session")
def cleanup():
    yield
    client = datastore.Client()
    query = client.query(kind=project_store)
    query_iter = query.fetch()
    for entity in query_iter:
        client.delete(entity.key)


@pytest.fixture(scope="session")
def cleanup_pdfslides():
    yield
    client = datastore.Client()
    query = client.query(kind=pdfslides_store)
    query_iter = query.fetch()
    for entity in query_iter:
        client.delete(entity.key)


@pytest.fixture(scope="session")
def cleanup_videosegment():
    yield
    client = datastore.Client()
    query = client.query(kind=videosegment_store)
    query_iter = query.fetch()
    for entity in query_iter:
        client.delete(entity.key)