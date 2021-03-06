module Main exposing (Model, Msg(..), init, main, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.CDN as CDN
import Bootstrap.Form as Form
import Bootstrap.Form.Textarea as Textarea
import Bootstrap.Grid as Grid
import Bootstrap.Navbar as Navbar
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Browser
import Browser.Navigation as Nav
import Css.Global exposing (path)
import File exposing (File)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as Decode exposing (Decoder, decodeString, float, int, list, null, string)
import Json.Decode.Pipeline exposing (hardcoded, optional, optionalAt, required, requiredAt)
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
    }


type Page
    = Index
    | Dashboard (Maybe String)
    | Login
    | Logout
    | Project String
    | Projects


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
        ]


init : Flags -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    let
        ( navbarState, navbarCmd ) =
            Navbar.initialState NavbarMsg
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
                            ( Model key url (urlToPage url) navbarState [] "" Alert.closed flags "", Cmd.batch [ navbarCmd, Nav.pushUrl key (Url.toString loginURL) ] )

                        Just userToken ->
                            ( Model key url (urlToPage url) navbarState [] "" Alert.closed flags userToken, Cmd.batch [ navbarCmd, Ports.storeToken userToken ] )

                Login ->
                    ( Model key url (urlToPage url) navbarState [] "" Alert.closed flags "", Cmd.batch [ navbarCmd ] )

                _ ->
                    ( Model key url (urlToPage url) navbarState [] "" Alert.closed flags "", Cmd.batch [ navbarCmd, Nav.pushUrl key (Url.toString loginURL) ] )

        Just userToken ->
            case urlToPage url of
                Project projectID ->
                    ( Model key url (urlToPage url) navbarState [] "" Alert.closed flags userToken, Cmd.batch [ navbarCmd ] )

                _ ->
                    ( Model key url (urlToPage url) navbarState [] "" Alert.closed flags userToken, Cmd.batch [ navbarCmd ] )



-- UPDATE


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | NavbarMsg Navbar.State
    | GotFiles (List File)
    | TemporaryResp (Result Http.Error String)
    | UpdateScriptTextArea String
    | SubmitJob
    | ToggleAlert Alert.Visibility
    | Tick Time.Posix


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
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

        TemporaryResp result ->
            case result of
                Ok items ->
                    ( model, Cmd.none )

                Err zzz ->
                    ( { model | alertVisibility = Alert.shown }, Cmd.none )

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
                            , Cmd.none
                            )

                        Project projectID ->
                            ( { model | url = url, page = urlToPage url }
                            , Cmd.none
                            )

                        Dashboard token ->
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
            , Grid.row []
                [ Grid.col []
                    [ case model.page of
                        Index ->
                            indexPage model.url.host model.url.path

                        Logout ->
                            indexPage "logout" "logout"

                        Login ->
                            let
                                dashboardURL =
                                    model.url
                            in
                            loginPage model.serverSettings.serverEndpoint { dashboardURL | path = model.serverSettings.ingressPath ++ "/dashboard" }

                        Projects ->
                            projectsPage

                        Project projectID ->
                            singleProjectPage

                        Dashboard token ->
                            dashboardPage
                    ]
                ]
            ]
        ]
    }


type alias SingleProject =
    { id : String
    , dateCreated : String
    , dateModified : String
    , status : String
    , videoOutputID : String
    }


singleProjectDecoder : Decoder SingleProject
singleProjectDecoder =
    Decode.succeed SingleProject
        |> required "id" string
        |> required "date_created" string
        |> required "date_modified" string
        |> required "status" string
        |> optional "video_output_id" string ""


indexPage : String -> String -> Html msg
indexPage aaa bbb =
    div [] [ text (aaa ++ bbb ++ "This is the Index Page. It is still not rendered out properly yet") ]


loginPage : String -> Url.Url -> Html msg
loginPage mgrurl sourceURL =
    div [] [ a [ href (mgrurl ++ "/api/v1/login?source_url=" ++ Url.toString sourceURL) ] [ text "Google Login" ] ]


dashboardPage : Html msg
dashboardPage =
    div [] [ text "Dashboard Page" ]


projectsPage : Html msg
projectsPage =
    div [] [ text "Projects Page" ]


singleProjectPage : Html msg
singleProjectPage =
    div [] [ text "Single Project Page" ]


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
