# ScreenLogic Wire Protocol

Byte order is little-endian.

## Data types

So far as I can tell, the following data types are present:

```
bool - encoded as a uint8 who's value is 1 or 0
uint8
uint16
uint32
String
raw bytes
```

The `String` type, the value is serialized length-encoded as follows:

|Field |Type          |
|------|--------------|
|Length|uint32        |
|Data  |[Length]bytes |

The `Data` portion is padded as 32bit segments. So, the string `"Spa"` would be encoded as:

```
\x03\x00\x00\x00Spa\x00
```

Notice the Length field value is `3`, but we need to be sure parse the trailing `\x00` as well in order to proceed with whatever comes next in the buffer. Just to be clear, the string `"Brian"` would be encoded as:

```
\x05\x00\x00\x00Brian\x00\x00\x00
```

This is just to show the trailing NULL bytes are indeed padding, not a C-string like the previous example appears to be.

## Framing

### Header

Packet frames have a 64bit header consisting of the following fields:

|Field      |Type  |
|-----------|------|
|Sequence   |uint16|
|MessageCode|uint16|
|BodyLength |uint32|

Following the header is the packet body.

There are two exceptions to this, which are the discovery packet response (which happens over UDP). And the initial string sent to the server on port 80 to initiate the connection. We'll go more into that below when describing the discovery and authentication sequence.

### Sequencing

As packets are sent back and forth from the client and server, the sequence field should increment after every request _and_ response. So, if for a request the sequence number was `2`, the response packet's sequence number will also be `2`. The next request packet sent should have a sequence value of `3`, and so on.