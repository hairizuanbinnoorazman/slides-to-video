- Add pubsub to shoot job ids
- Integrate microservices tgt -> its reports to this service
  - pdf splitter
  - image to video
  - video concatenate
- Allow to view list of jobs available
- Allow to download from a list of videos that are ready



{
    ID: UUID
    Filename: PDF name
    Script: Text
    Status: Not started, Running, Completed
    VideoFile: XXX
}

PDF to Image Job
{
    ID: UUID
    ParentJob: ID
    Status: Not started,Running, Completed
}

Image to video job
{
    ID: UUID
    ParentJob: ID
    Status: Not started, Running, Completed
    Output File: video file name
}

Video Concat Job
{
    ID: UUID
    ParentJob: ID
    Status: Not started, Running, Completed
    Output File: ID
}



curl -X POST http://localhost:8080/report_pdf_split -H "Content-Type: application/json" -d '{"id": "a443a907-091d-4767-b70b-c2be09824cc0", "status": "running"}'
curl -X POST http://localhost:8080/report_pdf_split -H "Content-Type: application/json" -d '{"id": "a443a907-091d-4767-b70b-c2be09824cc0", "status": "completed", "slide_details": [{"image": "1234.png", "slide_no": 0}, {"image": "2345.png", "slide_no": 1}]}'

curl -X POST http://localhost:8080/report_image_to_video -H "Content-Type: application/json" -d '{"id": "791ba4a4-b6b9-47ba-a23d-5994ab93ad42", "status":"running"}'
curl -X POST http://localhost:8080/report_image_to_video -H "Content-Type: application/json" -d '{"id": "d802c38a-735d-47b5-b7ad-871e3b4fa378", "status":"running"}'
curl -X POST http://localhost:8080/report_image_to_video -H "Content-Type: application/json" -d '{"id": "791ba4a4-b6b9-47ba-a23d-5994ab93ad42", "status":"completed", "output_file": "791ba4a4-b6b9-47ba-a23d-5994ab93ad42.mp4"}'
curl -X POST http://localhost:8080/report_image_to_video -H "Content-Type: application/json" -d '{"id": "d802c38a-735d-47b5-b7ad-871e3b4fa378", "status":"completed", "output_file": "d802c38a-735d-47b5-b7ad-871e3b4fa378.mp4"}'

curl -X POST http://localhost:8080/report_video_concat -H "Content-Type: application/json" -d '{"id": "4e0100ff-7cf2-4edb-bc82-2cb583f64aa9", "status":"running"}'
curl -X POST http://localhost:8080/report_video_concat -H "Content-Type: application/json" -d '{"id": "4e0100ff-7cf2-4edb-bc82-2cb583f64aa9", "status":"completed", "output_video": "db5ac55a-7913-49e5-bca4-d09c71ef0073.mp4"}'