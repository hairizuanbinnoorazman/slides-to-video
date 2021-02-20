# Slide to Video Manager

## Features to be developed

- User + ACL model integration (App-managed)
  - Require need to have the capability to limits to amount of projects that can be created per user
- Add Frontend for Integration - Plain HTML + Golang Server
- Setup monitoring (Prometheus Endpoints)
- Setup Distributed Tracing between components
- Setup Profiling Endpoints (switched on via configuration)
- Add documentation regarding the deployment to local (docker-compose), kubernetes, cloud run
- List operations for Project/VideoSegments/PDFSlideImages are broken
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
