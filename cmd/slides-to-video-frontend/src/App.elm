module App exposing (Flags, Model, Msg(..), init, main, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.CDN as CDN
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Textarea as Textarea
import Bootstrap.Grid as Grid
import Bootstrap.Navbar as Navbar
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Browser
import Browser.Navigation as Nav
import Bytes
import Bytes.Decode
import Css.Global exposing (path)
import File exposing (File)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http exposing (Error, Header)
import Image
import Json.Decode as Decode exposing (Decoder, bool, decodeString, float, int, list, null, string)
import Json.Decode.Pipeline as Pipeline
import Json.Encode as Encode
import Ports
import String
import Time
import Url
import Url.Parser as Url exposing ((</>), (<?>), Parser)
import Url.Parser.Query as Query



-- MAIN


type alias Flags =
    { serverEndpoint : String
    , ingressPath : String
    , token : Maybe String
    }


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }



-- MODEL


type alias Model =
    { key : Nav.Key
    , url : Url.Url
    , page : Page
    , navbarState : Navbar.State
    , files : List File
    , script : String
    , alertVisibility : Alert.Visibility
    , serverSettings : Flags
    , userToken : String
    , userDetails : UserDetails
    , projects : ProjectList
    , singleProject : SingleProject
    }


type alias UserDetails =
    { username : String
    , password : String
    , passwordAgain : String
    }


type Page
    = Index
    | Dashboard (Maybe String)
    | Login
    | Logout
    | Project String
    | Projects
    | UserRegister


urlToPage : Url.Url -> Page
urlToPage url =
    url
        |> Url.parse urlParser
        |> Maybe.withDefault Index


urlParser : Parser (Page -> a) a
urlParser =
    -- We try to match one of the following URLs
    Url.oneOf
        [ Url.map Index Url.top
        , Url.map Login (Url.s "login")
        , Url.map Logout (Url.s "logout")
        , Url.map Dashboard (Url.s "dashboard" <?> Query.string "token")
        , Url.map Projects (Url.s "projects")
        , Url.map Project (Url.s "projects" </> Url.string)
        , Url.map UserRegister (Url.s "register")
        ]


init : Flags -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    let
        ( navbarState, navbarCmd ) =
            Navbar.initialState NavbarMsg

        emptyProjectList =
            ProjectList [] 0 0 0

        emptyUserDetails =
            UserDetails "" "" ""

        emptySingleProject =
            SingleProject "" "" "" "" "" [] [] ""

        initialAppState =
            { key = key
            , url = url
            , page = urlToPage url
            , navbarState = navbarState
            , files = []
            , script = ""
            , alertVisibility = Alert.closed
            , serverSettings = flags
            , userToken = ""
            , userDetails = emptyUserDetails
            , projects = emptyProjectList
            , singleProject = emptySingleProject
            }
    in
    case flags.token of
        Nothing ->
            let
                loginURL =
                    { url | path = "/login" }
            in
            case urlToPage url of
                Dashboard newUserToken ->
                    case newUserToken of
                        Nothing ->
                            ( initialAppState, Cmd.batch [ navbarCmd, Nav.pushUrl key (Url.toString loginURL) ] )

                        Just userToken ->
                            ( { initialAppState | userToken = userToken }, Cmd.batch [ navbarCmd, Ports.storeToken userToken ] )

                Login ->
                    ( initialAppState, Cmd.batch [ navbarCmd ] )

                UserRegister ->
                    ( initialAppState, Cmd.batch [ navbarCmd ] )

                _ ->
                    ( initialAppState, Cmd.batch [ navbarCmd, Nav.pushUrl key (Url.toString loginURL) ] )

        Just userToken ->
            case urlToPage url of
                Project projectID ->
                    ( { initialAppState | userToken = userToken }, Cmd.batch [ navbarCmd ] )

                _ ->
                    ( { initialAppState | userToken = userToken }, Cmd.batch [ navbarCmd ] )



-- UPDATE


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | NavbarMsg Navbar.State
    | GotFiles (List File)
    | TemporaryResp (Result Http.Error String)
    | EmptyRedirectResponse (Result Http.Error ())
    | EmptyResponse (Result Http.Error ())
    | LoginResponse (Result Http.Error ())
    | ProjectsResponse (Result Http.Error ProjectList)
    | UpdateScriptTextArea String
    | SubmitJob
    | ToggleAlert Alert.Visibility
    | Tick Time.Posix
    | UsernameInput String
    | PasswordInput String
    | PasswordAgainInput String
    | RegisterUserCredentials
    | SubmitLoginCredentials
    | CreateNewProject
    | CreateProjectResponse (Result Http.Error SingleProject)
    | GetProjectResponse (Result Http.Error SingleProject)
    | GotImage (Result Http.Error (Maybe Image.Image))


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ProjectsResponse result ->
            case result of
                Ok zzz ->
                    ( { model | projects = zzz }, Cmd.none )

                Err zzz ->
                    ( model, Cmd.none )

        LoginResponse result ->
            case result of
                Ok zzz ->
                    -- ( { model | userToken = zzz.token }, Cmd.batch [ Ports.storeToken zzz.token, Nav.pushUrl model.key "/" ] )
                    ( { model | userToken = "fake" }, Cmd.batch [ Nav.pushUrl model.key "/" ] )

                Err zzz ->
                    ( { model | alertVisibility = Alert.shown }, Cmd.none )

        SubmitLoginCredentials ->
            let
                tempUsername =
                    model.userDetails.username

                tempPassword =
                    model.userDetails.password
            in
            ( { model | userDetails = UserDetails "" "" "" }, Cmd.batch [ loginUser model.serverSettings.serverEndpoint tempUsername tempPassword ] )

        RegisterUserCredentials ->
            let
                tempUsername =
                    model.userDetails.username

                tempPassword =
                    model.userDetails.password
            in
            ( { model | userDetails = UserDetails "" "" "" }, Cmd.batch [ createUser model.serverSettings.serverEndpoint tempUsername tempPassword ] )

        UsernameInput username ->
            ( { model | userDetails = UserDetails username model.userDetails.password model.userDetails.passwordAgain }, Cmd.none )

        PasswordInput password ->
            ( { model | userDetails = UserDetails model.userDetails.username password model.userDetails.passwordAgain }, Cmd.none )

        PasswordAgainInput passwordAgain ->
            ( { model | userDetails = UserDetails model.userDetails.username model.userDetails.password passwordAgain }, Cmd.none )

        Tick time ->
            case model.page of
                _ ->
                    ( model, Cmd.none )

        ToggleAlert alertVisibility ->
            ( { model | alertVisibility = alertVisibility }, Cmd.none )

        SubmitJob ->
            ( model, Cmd.batch [ uploadFile model.serverSettings.serverEndpoint model.files ] )

        UpdateScriptTextArea scriptText ->
            ( { model | script = scriptText }, Cmd.none )

        EmptyRedirectResponse result ->
            case result of
                Ok a ->
                    ( model, Nav.pushUrl model.key "/" )

                Err a ->
                    ( { model | alertVisibility = Alert.shown }, Cmd.none )

        EmptyResponse result ->
            case result of
                Ok a ->
                    ( model, Cmd.none )

                Err a ->
                    ( { model | alertVisibility = Alert.shown }, Cmd.none )

        CreateProjectResponse result ->
            case result of
                Ok p ->
                    ( { model | singleProject = p }, Nav.pushUrl model.key ("/projects/" ++ p.id) )

                Err _ ->
                    ( { model | alertVisibility = Alert.shown }, Cmd.none )

        TemporaryResp result ->
            case result of
                Ok items ->
                    ( model, Cmd.none )

                Err zzz ->
                    ( { model | alertVisibility = Alert.shown }, Cmd.none )

        GotImage result ->
            case result of
                Ok items ->
                    ( model, Cmd.none )

                Err zzz ->
                    ( model, Cmd.none )

        GotFiles files ->
            ( { model | files = files }, Cmd.none )

        NavbarMsg state ->
            ( { model | navbarState = state }, Cmd.none )

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        CreateNewProject ->
            ( model, apiCreateProject model.serverSettings.serverEndpoint model.userToken )

        GetProjectResponse result ->
            case result of
                Ok p ->
                    ( { model | singleProject = p }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        UrlChanged url ->
            case model.userToken of
                "" ->
                    let
                        loginURL =
                            { url | path = "/login" }
                    in
                    case urlToPage url of
                        Login ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )

                        UserRegister ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )

                        _ ->
                            ( model
                            , Cmd.batch [ Nav.pushUrl model.key (Url.toString loginURL) ]
                            )

                _ ->
                    let
                        indexURL =
                            { url | path = model.serverSettings.ingressPath ++ "/" }
                    in
                    case urlToPage url of
                        Index ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )

                        Login ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )

                        Logout ->
                            ( { model | url = url, page = urlToPage url, userToken = "" }
                            , Cmd.batch [ Nav.pushUrl model.key (Url.toString indexURL), Ports.removeToken () ]
                            )

                        Projects ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.batch [ apiListProjects model.serverSettings.serverEndpoint model.userToken ]
                            )

                        Project projectID ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.batch
                                [ apiGetProject model.serverSettings.serverEndpoint model.userToken projectID
                                ]
                            )

                        Dashboard token ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )

                        UserRegister ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )


errorToString : Http.Error -> String
errorToString error =
    case error of
        Http.BadUrl url ->
            "The URL " ++ url ++ " was invalid"

        Http.Timeout ->
            "Unable to reach the server, try again"

        Http.NetworkError ->
            "Unable to reach the server, check your network connection"

        Http.BadStatus 500 ->
            "The server had a problem, try again later"

        Http.BadStatus 400 ->
            "Verify your information and try again"

        Http.BadStatus _ ->
            "Unknown error"

        Http.BadBody errorMessage ->
            errorMessage



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Time.every 10000 Tick



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Slides To Video Creation Tool"
    , body =
        [ Grid.container []
            [ CDN.stylesheet -- creates an inline style node with the Bootstrap CSS
            , Grid.row []
                [ Grid.col []
                    [ Navbar.config NavbarMsg
                        |> Navbar.withAnimation
                        |> Navbar.brand [ href (model.serverSettings.ingressPath ++ "/dashboard") ] [ text "Brand" ]
                        |> Navbar.items
                            [ Navbar.itemLink [ href (model.serverSettings.ingressPath ++ "/dashboard") ] [ text "Dashboard" ]
                            , Navbar.itemLink [ href (model.serverSettings.ingressPath ++ "/projects") ] [ text "Projects" ]
                            ]
                        |> Navbar.customItems
                            [ case model.userToken of
                                "" ->
                                    Navbar.customItem (a [ class "nav-link", href (model.serverSettings.ingressPath ++ "/login") ] [ text "Login" ])

                                _ ->
                                    Navbar.customItem (a [ class "nav-link", href (model.serverSettings.ingressPath ++ "/logout") ] [ text "Logout" ])
                            ]
                        |> Navbar.view model.navbarState
                    ]
                ]
            , case model.page of
                Index ->
                    indexPage model.url.host model.url.path

                Logout ->
                    indexPage "logout" "logout"

                Login ->
                    let
                        dashboardURL =
                            model.url
                    in
                    loginPage model { dashboardURL | path = model.serverSettings.ingressPath ++ "/dashboard" }

                Projects ->
                    projectsPage model

                Project projectID ->
                    singleProjectPage model.singleProject

                Dashboard token ->
                    dashboardPage

                UserRegister ->
                    registerPage model
            ]
        ]
    }


type alias UserToken =
    { token : String
    }


userTokenDecoder : Decoder UserToken
userTokenDecoder =
    Decode.succeed UserToken
        |> Pipeline.required "token" string


type alias SingleProject =
    { id : String
    , name : String
    , dateCreated : String
    , dateModified : String
    , status : String
    , pdfSlideImages : List PDFSlideImages
    , videoSegments : List VideoSegment
    , videoOutputID : String
    }


singleProjectDecoder : Decoder SingleProject
singleProjectDecoder =
    Decode.succeed SingleProject
        |> Pipeline.required "id" string
        |> Pipeline.required "name" string
        |> Pipeline.required "date_created" string
        |> Pipeline.required "date_modified" string
        |> Pipeline.required "status" string
        |> Pipeline.optional "pdf_slide_images" (Decode.list pdfSlideImagesDecoder) []
        |> Pipeline.optional "video_segments" (Decode.list videoSegmentDecoder) []
        |> Pipeline.optional "video_output_id" string ""


type alias ProjectList =
    { projects : List SingleProject
    , limit : Int
    , offset : Int
    , total : Int
    }


projectListDecoder : Decoder ProjectList
projectListDecoder =
    Decode.succeed ProjectList
        |> Pipeline.required "projects" (Decode.list singleProjectDecoder)
        |> Pipeline.required "limit" int
        |> Pipeline.required "offset" int
        |> Pipeline.required "total" int


type alias SlideAsset =
    { imageID : String
    , order : Int
    }


slideAssetDecoder : Decoder SlideAsset
slideAssetDecoder =
    Decode.succeed SlideAsset
        |> Pipeline.required "image_id" string
        |> Pipeline.required "order" int


type alias PDFSlideImages =
    { id : String
    , projectID : String
    , pdfFile : String
    , dateCreated : String
    , slideAssets : List SlideAsset
    , status : String
    }


pdfSlideImagesDecoder : Decoder PDFSlideImages
pdfSlideImagesDecoder =
    Decode.succeed PDFSlideImages
        |> Pipeline.required "id" string
        |> Pipeline.required "project_id" string
        |> Pipeline.required "pdf_file" string
        |> Pipeline.required "date_created" string
        |> Pipeline.required "slide_assets" (Decode.list slideAssetDecoder)
        |> Pipeline.required "status" string


type alias VideoSegment =
    { id : String
    , projectID : String
    , videoFile : String
    , dateCreated : String
    , dateModified : String
    , order : Int
    , hidden : Bool
    , status : String
    , imageID : String
    , script : String
    , audioID : String
    , videoSrcID : String
    }


videoSegmentDecoder : Decoder VideoSegment
videoSegmentDecoder =
    Decode.succeed VideoSegment
        |> Pipeline.required "id" string
        |> Pipeline.required "project_id" string
        |> Pipeline.required "video_file" string
        |> Pipeline.required "date_created" string
        |> Pipeline.required "date_modified" string
        |> Pipeline.required "order" int
        |> Pipeline.required "hidden" bool
        |> Pipeline.required "status" string
        |> Pipeline.optional "image_id" string ""
        |> Pipeline.optional "script" string ""
        |> Pipeline.optional "audio_id" string ""
        |> Pipeline.optional "video_src_id" string ""


loginPage : Model -> Url.Url -> Html Msg
loginPage model sourceURL =
    Grid.row []
        [ Grid.col []
            [ Alert.config
                |> Alert.danger
                |> Alert.dismissable ToggleAlert
                |> Alert.children
                    [ p [] [ text "Unable to login" ]
                    ]
                |> Alert.view model.alertVisibility
            , h2 [] [ text "Login" ]
            , Form.form []
                [ Form.group []
                    [ Form.label [ for "useremail" ] [ text "Email address" ]
                    , Input.email [ Input.id "useremail", Input.value model.userDetails.username, Input.onInput UsernameInput ]
                    , Form.help [] [ text "We'll never share your email with anyone else." ]
                    ]
                , Form.group []
                    [ Form.label [ for "userpassword" ] [ text "Password" ]
                    , Input.password [ Input.id "userpassword", Input.value model.userDetails.password, Input.onInput PasswordInput ]
                    ]
                , Button.button [ Button.primary, Button.onClick SubmitLoginCredentials ] [ text "Login" ]
                ]
            , a [ href (model.serverSettings.serverEndpoint ++ "/api/v1/login?source_url=" ++ Url.toString sourceURL) ] [ text "Google Login" ]
            , br [] []
            , a [ href "/register" ] [ text "Register with Email" ]
            ]
        ]


registerPage : Model -> Html Msg
registerPage model =
    Grid.row []
        [ Grid.col []
            [ Alert.config
                |> Alert.danger
                |> Alert.dismissable ToggleAlert
                |> Alert.children
                    [ p [] [ text "Unable to register user" ]
                    ]
                |> Alert.view model.alertVisibility
            , h2 [] [ text "Register New Account" ]
            , Form.form []
                [ Form.group []
                    [ Form.label [ for "useremail" ] [ text "Email address" ]
                    , Input.email [ Input.id "useremail", Input.value model.userDetails.username, Input.onInput UsernameInput ]
                    , Form.help [] [ text "We'll never share your email with anyone else." ]
                    ]
                , Form.group []
                    [ Form.label [ for "userpassword" ] [ text "Password" ]
                    , Input.password [ Input.id "userpassword", Input.value model.userDetails.password, Input.onInput PasswordInput ]
                    ]
                ]
            , Form.group []
                [ Form.label [ for "confirmuserpassword" ] [ text "Confirm Password" ]
                , Input.password [ Input.id "confirmuserpassword", Input.value model.userDetails.passwordAgain, Input.onInput PasswordAgainInput ]
                ]
            , if model.userDetails.password == model.userDetails.passwordAgain then
                div []
                    [ p [ style "color" "green" ] [ text "OK" ]
                    , Button.button [ Button.primary, Button.onClick RegisterUserCredentials ] [ text "Submit" ]
                    ]

              else
                p [ style "color" "red" ] [ text "Passwords do not match!" ]
            ]
        ]


dashboardPage : Html Msg
dashboardPage =
    div [] [ h1 [] [ text "Dashboard Page" ] ]


indexPage : String -> String -> Html Msg
indexPage aaa bbb =
    div [] [ text (aaa ++ bbb ++ "This is the Index Page. It is still not rendered out properly yet") ]


singleProjectRow : SingleProject -> Table.Row msg
singleProjectRow singleProject =
    Table.tr []
        [ Table.td [] [ a [ href ("/projects/" ++ singleProject.id) ] [ text singleProject.name ] ]
        , Table.td [] [ text singleProject.dateCreated ]
        , Table.td [] [ text singleProject.dateModified ]
        , Table.td [] [ text singleProject.status ]
        , Table.td []
            [ if singleProject.status == "completed" then
                a [] [ text "Download Link" ]

              else
                p [] [ text "Not available" ]
            ]
        ]


projectsPage : Model -> Html Msg
projectsPage model =
    Grid.row []
        [ Grid.col []
            [ Alert.config
                |> Alert.danger
                |> Alert.dismissable ToggleAlert
                |> Alert.children
                    [ p [] [ text "Unable to fetch projects list" ]
                    ]
                |> Alert.view model.alertVisibility
            , h2 [] [ text "Projects" ]
            , Button.button [ Button.primary, Button.onClick CreateNewProject ] [ text "Create Project" ]
            , if List.length model.projects.projects == 0 then
                p [] [ text "No projects found" ]

              else
                Table.simpleTable
                    ( Table.simpleThead
                        [ Table.th [] [ text "Name" ]
                        , Table.th [] [ text "Date Created" ]
                        , Table.th [] [ text "Last Modified" ]
                        , Table.th [] [ text "Status" ]
                        , Table.th [] [ text "Video Download Link" ]
                        ]
                    , Table.tbody []
                        (List.map singleProjectRow model.projects.projects)
                    )
            ]
        ]


singleProjectPage : SingleProject -> Html Msg
singleProjectPage singleProject =
    div []
        (List.concat
            [ [ h1 [] [ text "Project" ]
              , Button.button [ Button.primary ] [ text "Save" ]
              , if List.length singleProject.pdfSlideImages == 0 then
                    p [] [ text "Please upload a PDF File containing the slides" ]

                else
                    p [] [ text "You may replace the PDF File" ]
              , input [ type_ "file", multiple False, on "change" (Decode.map GotFiles filesDecoder) ] []
              ]
            , List.map videoSegmentRow singleProject.videoSegments
            ]
        )


videoSegmentRow : VideoSegment -> Html Msg
videoSegmentRow videoSegment =
    Card.config []
        |> Card.block []
            [ Block.custom <|
                Form.form []
                    [ Form.group []
                        [ div []
                            [ p [] [ text ("Possible location of image :: " ++ videoSegment.imageID) ]
                            , p [] [ text "Possible location of video" ]
                            ]
                        , Form.label [] [ text "Enter script for segment" ]
                        , Textarea.textarea
                            [ Textarea.id "script"
                            , Textarea.rows 3
                            , Textarea.value videoSegment.script
                            ]
                        ]
                    , Form.group []
                        [ Button.button [ Button.primary ] [ text "Save script" ]
                        , Button.button [ Button.primary ] [ text "Generate Audio" ]
                        , Button.button [ Button.primary ] [ text "Move up" ]
                        , Button.button [ Button.primary ] [ text "Move down" ]
                        ]
                    ]
            ]
        |> Card.view


filesDecoder : Decoder (List File)
filesDecoder =
    Decode.at [ "target", "files" ] (Decode.list File.decoder)


uploadFile : String -> List File -> Cmd Msg
uploadFile mgrURL files =
    Http.post
        { url = mgrURL ++ "/api/v1/job"
        , expect = Http.expectString TemporaryResp
        , body = Http.multipartBody (List.map (Http.filePart "myfile") files)
        }


createUser : String -> String -> String -> Cmd Msg
createUser mgrURL userEmail userPassword =
    let
        url =
            mgrURL ++ "/api/v1/users/register"

        body =
            Http.jsonBody <|
                Encode.object
                    [ ( "email", Encode.string userEmail )
                    , ( "password", Encode.string userPassword )
                    ]
    in
    Http.request
        { body = body
        , method = "POST"
        , url = url
        , headers = []
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectWhatever EmptyRedirectResponse
        }


apiListProjects : String -> String -> Cmd Msg
apiListProjects mgrURL apiToken =
    let
        url =
            mgrURL ++ "/api/v1/projects"
    in
    Http.request
        { body = Http.emptyBody
        , method = "GET"
        , url = url
        , headers =
            [ Http.header "Authorization" ("Bearer " ++ apiToken)
            ]
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectJson ProjectsResponse projectListDecoder
        }


apiCreateProject : String -> String -> Cmd Msg
apiCreateProject mgrURL apiToken =
    let
        url =
            mgrURL ++ "/api/v1/project"
    in
    Http.request
        { body = Http.emptyBody
        , method = "POST"
        , url = url
        , headers =
            [ Http.header "Authorization" ("Bearer " ++ apiToken)
            ]
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectJson CreateProjectResponse singleProjectDecoder
        }


apiGetProject : String -> String -> String -> Cmd Msg
apiGetProject mgrURL apiToken projectID =
    let
        url =
            mgrURL ++ "/api/v1/project/" ++ projectID
    in
    Http.request
        { body = Http.emptyBody
        , method = "GET"
        , url = url
        , headers =
            [ Http.header "Authorization" ("Bearer " ++ apiToken)
            ]
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectJson GetProjectResponse singleProjectDecoder
        }


apiUpdateProject : String -> String -> String -> String -> Cmd Msg
apiUpdateProject mgrURL apiToken projectID projectName =
    let
        url =
            mgrURL ++ "/api/v1/project/" ++ projectID

        body =
            Http.jsonBody <|
                Encode.object
                    [ ( "name", Encode.string projectName )
                    ]
    in
    Http.request
        { body = body
        , method = "PUT"
        , url = url
        , headers =
            [ Http.header "Authorization" ("Bearer " ++ apiToken)
            ]
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectJson CreateProjectResponse singleProjectDecoder
        }


apiUploadPDFSlides : String -> String -> String -> List File -> Cmd Msg
apiUploadPDFSlides mgrURL apiToken projectID files =
    let
        url =
            mgrURL ++ "/api/v1/project/" ++ projectID ++ "/pdfslideimages"
    in
    Http.request
        { body = Http.multipartBody (List.map (Http.filePart "myfile") files)
        , method = "POST"
        , url = url
        , headers =
            [ Http.header "Authorization" ("Bearer " ++ apiToken)
            ]
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectJson CreateProjectResponse singleProjectDecoder
        }


type alias FullImageResponse =
    { imageID : String
    , image : Maybe Image.Image
    }


decodeBytesToImage : Http.Response Bytes.Bytes -> Result Http.Error (Maybe Image.Image)
decodeBytesToImage response =
    case response of
        Http.GoodStatus_ _ body ->
            Ok <| Image.decode body

        Http.BadUrl_ url ->
            Err (Http.BadUrl url)

        Http.Timeout_ ->
            Err Http.Timeout

        Http.NetworkError_ ->
            Err Http.NetworkError

        Http.BadStatus_ metadata _ ->
            Err (Http.BadStatus metadata.statusCode)


moddedAPIGetImageAsset : String -> String -> String -> (String -> Cmd Msg)
moddedAPIGetImageAsset mgrURL apiToken projectID =
    apiGetImageAsset mgrURL apiToken projectID


apiGetImageAsset : String -> String -> String -> String -> Cmd Msg
apiGetImageAsset mgrURL apiToken projectID imageID =
    let
        url =
            mgrURL ++ "/api/v1/project/" ++ projectID ++ "/image/" ++ imageID
    in
    Http.request
        { body = Http.emptyBody
        , method = "GET"
        , url = url
        , headers =
            [ Http.header "Authorization" ("Bearer " ++ apiToken)
            ]
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectBytesResponse GotImage decodeBytesToImage
        }


loginUser : String -> String -> String -> Cmd Msg
loginUser mgrURL userEmail userPassword =
    let
        url =
            mgrURL ++ "/api/v1/login"

        body =
            Http.jsonBody <|
                Encode.object
                    [ ( "email", Encode.string userEmail )
                    , ( "password", Encode.string userPassword )
                    ]
    in
    Http.request
        { body = body
        , method = "POST"
        , url = url
        , headers = []
        , timeout = Nothing
        , tracker = Nothing
        , expect = Http.expectWhatever LoginResponse
        }
