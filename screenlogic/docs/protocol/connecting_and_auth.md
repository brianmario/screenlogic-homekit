# ScreenLogic Wire Protocol - Connecting and Authentication

## Discover Gateways on the Network

By default the gateway is configured to join the network with DHCP, so this step is most likely required.

If you don't already know the connection info for the gateway(s) on your network, you'll need to perform a UDP broadcast on port 1444, with a [Discovery Request](types.md#Discovery) packet as the payload.

The gateway(s) will respond with a [Discovery Response](types.md#Discovery) packet, which contains the IP and port of the gateway. You'll use that to directly connect over IP.

## Connection

Once you've established a connection with the gateway over IP, you'll immediately send the [Challenge Request](types.md#Challenge) packet pair.

The gateway will then respond with a [Challenge Response](types.md#Challenge) packet, which contains the mac address of the gateway. From what I can tell looking elsewhere, it's possible the mac address is somehow used during password encryption or something. I'm not sure yet.

## Authentication

After the connection's challenge sequence is complete, you then send a [Login Request](types.md#Login) packet. I'm not sure what most of the fields are used for, but this is also where you would pass the encrypted password.

If authenticating locally (on the same network), you don't actually need to specify a password and can instead leave that field blank (16 `NULL` byes). If logging in remotely I'm pretty sure a password is required, though I haven't tried that. The encryption used for the password is something I don't recognize.

Once authentication is accepted by the gateway, a [Login Response](types.md#Login) packet will be sent back.

If authentication fails, a [Login Failed](types.md#Login%20Failed) packet will be returned.

## Errors

At the time of writing, there are only two error packet types that I know of so far. A Login Failed packet and a Bad Parameter packet.

[Login Failed](types.md#Login%20Failed) is obivously returned if authentication fails.

[Bad Parameter](types.md#Bad%20Parameter) is I think used if an unknown packet header `Code` value is used. So basically, an unknown command.