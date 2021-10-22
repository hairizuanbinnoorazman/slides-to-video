module Reactor exposing (..)

import App exposing (..)
import Browser
import Url
import Browser.Navigation as Nav

rinit : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
rinit () url key =
    init {serverEndpoint = "http://localhost:8081", ingressPath = "", token = Nothing} url key

main : Program () Model Msg
main =
    Browser.application
        { init = rinit
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }