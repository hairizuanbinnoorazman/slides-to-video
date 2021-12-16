# Slide to Video Manager

## Quickstart

For quickly getting started with this on local environment. Current minimum requirements is 4 core and 4.5 GB of memory

```bash
make build-bin
make build-images
make stack-up
```

To test that all endpoints is working:

```bash
cd ./tests
pipenv shell
pipenv install
pytest test_app.py
```

To test frontend + develop frontend

```bash
elm reactor
# Head to Reactor.elm -> Default values is set 

# Before commiting to git repo - ensure that elm code is formatted well
make format
```

## Features to be developed

- Setup docker compose with frontend
  - Require transpiling as well as uglify of elm codebase
- Provide a user page with some details
  - Create dashboard on per user basis
  - Maximum no of projects available for user
  - Current usage of number of projects for user
- Videos located in the wrong folder in minio S3 -> should be configurable
- Access Image and Video assets
  - Image for video segment -> Need to check that the user can actually really access the image (it has cookie protection)
  - Audio for video segment
- Projects
  - How to handle status for when no "long running" process is happening
  - Count of number of projects by user?
  - If there is a long running pdf concater - do not enable the capability to generate?
  - Roles
    - Owner - Can be group/user (For project accounting)
    - Publisher - Send end result to a video streaming site
    - Editor - Can edit project, export project
    - Viewer - Can view project
- User + ACL model integration (App-managed)
  - Require need to have the capability to limits to amount of projects that can be created per user
  - Rate limiting of all APIs
- API Security
  - Add capability to check presence of authorization -> have flag to disable it for integration testing purposes
- Build out a simple user authentication system that can be used for integration testing purposes
  - Passwordless based (Send email links for login)
- Integrate Email testing for integration testing purposes
  - https://mailslurper.com/
  - https://github.com/inbucket
- Groups API
  - Multiple Users can be in a group
  - Roles
    - Owner (Can invite other members, View Project Count Limit, Create Projects, Delete Projects, Shut down group, Export all project assets, Archive Projects)
    - Member
- Projects ACL
  - Introduce Tagging to allow easier filtering (Maybe next time)
- Setup monitoring (Prometheus Endpoints)
- Setup Distributed Tracing between components
- Setup Profiling Endpoints (switched on via configuration)
- Add documentation regarding the deployment to local (docker-compose), kubernetes, cloud run
- List operations for VideoSegments/PDFSlideImages are broken
- Delete operations for Project/VideoSegments/PDFSlideImages are broken
- Support of cassandra as alternative data storage
- Support of local storage as alternative "blob storage"
- Support of kafka as queue system
- Support of rabbitmq as queue system
- Support deployment mode into GKE (API server - utilizes Google Datastore + Workers - utilizes Google Pubsub Pull mode)
- Move migrate command to utilize this: https://github.com/golang-migrate/migrate - Automigrate now is only meant to get the initial scheme into db
- User + ACL model integration (Using Keycloak)
- Add search capability (Elasticsearch)
- Add search capability (Bleve)

# Weird Things to take note of

- Downloading of images before showing  
  https://github.com/justgook/elm-image/issues/9
- Handling repeated items elm  
  https://stackoverflow.com/questions/48734554/view-with-multiple-textareas-how-to-identify-which-was-modified

## Development workflow

The application here are aimed to be deployed via Cloud Run and Kubernetes as primary deployment targets. In order to make it slightly easier to deploy apps to K8s clusters, skaffold tool is being used as part of development workflow to quickly test integration/configuration changes of helm chart/dockerfiles

## Happy Path

This defines expected userflow when using the tool:

- User uploads PDF file to platform
- Loads to Project page
- Project page loads up all pictures after a while of processing
- User adds in text in each of the slides
- User clicks to generate the video at the end
- Video link would be made available at the top of Single Project Page

Api flow - based on above flow

- Create project (accepts PDF file) -> should return a Project ID -> create project should create PDF Split Job
- Frontend goes to project page -> calls GetProject (ID)
- User refreshes manually (maybe future flow can adjust) -> call GetProject (ID) -> images should update soon
- After user adds text -> can click save at the top -> call PatchProject (ID) -> only script text can be updated here
- After user is done with adding text -> click on GenerateVideo -> calls to Project:Generate API -> this one would create the image to video jobs as well as video concat jobs accordingly

## Planning - Future features

Main interaction

User

- ID - uuid
- Name
- DateCreated
- DateModified
- GoogleRefreshToken
- GoogleAccessToken
- GoogleAccessExpiry

Group

- ID - uuid
- Name
- DateCreated
- DateModified
- UserIDs (List)
- GroupIDs (List) - max - 5 nested

Project

- ID
- Name
- Tags
- DateCreated
- DateModified
- UserAccess (List)
  - UserID
  - GroupID
- FinalVideo
  - OutputFile
  - Status
  - DateTimeRequest
- VideoSegments (List)
  - VideoFile
  - DateCreated
  - Status
  - DateTimeRequest
  - Order
  - ImageSource
    - ImageID
    - Type (Google Slides, Raw, PDF)
  - AudioSource
    - AudioID
    - Type (Google Source, Raw)
  - VideoSource (VideoSources takes priority)
    - VideoID (VideoFile == VideoID)
    - Type (Raw)
- ImageImporters (List)
  - PDF (No update operation allowed)
    - PDFFile
    - DateCreated
    - Slides (List)
      - SlideImage
      - SlideOrder
  - GoogleSlides (No update operation allowed)
    - ID
    - DateCreated
    - SlideURL
    - Slides (List)
      - SlideImage
      - SlideOrder
  - GoogleSlidesVersioned
    - ID
    - DateCreated
    - SlideURL
    - UserID
    - SlideVersion
      - ID
      - DateCreated
      - SlideCount
      - Slides (List)
        - SlideImage
        - SlideOrder
  - PDFVersioned
    - ID
    - PDFVersion
      - ID
      - DateCreated
      - SlideCount
      - Slides (List)
        - SlideImage
        - SlideOrder
