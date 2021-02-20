port module Ports exposing (removeToken, storeToken)


port storeToken : String -> Cmd msg


port removeToken : () -> Cmd msg
