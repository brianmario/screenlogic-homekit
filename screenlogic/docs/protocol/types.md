# ScreenLogic Wire Protocol - Packet Types

## Discovery

### Request

This packet consists of a frame header only. The `Sequence` field must be `1`, presumably because it's the first packet in the sequence ;)

The packet should look as follows over the wire:

```
\x01\x00\x00\x00\x00\x00\x00\x00
```

### Response

NOTE: This packet has no frame header, it is sent exactly as described below.

|Field        |Type   |
|-------------|-------|
|Type         |uint32 |
|IPAddr       |[4]byte|
|Port         |uint16 |
|GatewayType  |uint8  |
|GatewaySubnet|uint8  |
|GatewayName  |String |

## Challenge

This is the first set of packets sent after direct TCP/IP connection is established with the gateway.

### Request

There are technically two request "packets", though one of them is more of a HTTP-formatted command string.

The first is sent immediately after connecting to the gateway, and is as follows:

```
CONNECTSERVERHOST\r\n\r\n
```

Right after that, a Challenge message packet should be sent. This packet is a header only, who's `Code` field is `14`.

### Response

The response packet consists of a single field, which is the mac address of the gateway.

|Field         |Type   |
|--------------|-------|
|GatewayMacAddr|String |

## Login

### Request

This packet's `Code` field should be `27`.

At the time of writing, this library hasn't implemented the encryption needed to send passwords for authentication. So for now, if you're using this library, the password should be left blank.

This also means you'll only be able to make local connections (remote connection functionality hasn't been implemented either). And being as though this library is meant to be used with HomeKit, and HomeKit itself will allow remote access, I may not ever implement remote connection support. Just FYI.

|Field         |Type    |
|--------------|--------|
|Schema        |uint32  |
|ConnectionType|uint32  |
|ClientName    |String  |
|Password      |[16]byte|
|PID           |uint32  |

### Response

The response packet will consist of a header only, and the `Code` field is `28`.

## Get Gateway Version

### Request

This packet consists of a header only, who's `Code` field is `8120`.

### Response

This packet consists of a header only, who's `Code` field is `8121`.

|Field        |Type  |
|-------------|------|
|Version      |String|
|UnknownField1|uint32|
|UnknownField2|uint32|
|UnknownField3|uint32|
|UnknownField4|uint32|
|UnknownField5|uint32|
|UnknownField6|uint32|

The `Version` field is the gateway firmware version.

The rest of the fields, I have no idea.

## Get Gateway Configuration

### Request

This packet header's `Code` field is `12532`.

The body consists of the following fields:

|Field        |Type  |
|-------------|------|
|UnknownField1|uint32|
|UnknownField2|uint32|

I have no idea what they're for.

### Response

This packet header's `Code` field is `12533`.

The `Circuit`, `Color` and `Pump` types referenced here are described just below.

|Field             |Type                |
|------------------|--------------------|
|ControllerID      |uint32              |
|SetPointMin0      |uint8               |
|SetPointMax0      |uint8               |
|SetPointMin1      |uint8               |
|SetPointMax1      |uint8               |
|IsCelcius         |bool                |
|ControllerType    |uint8               |
|HardwareType      |uint8               |
|ControllerBuffer  |uint8               |
|EquipmentFlags    |uint32              |
|DefaultCircuitName|String              |
|NumCircuits       |uint32              |
|Circuits          |[NumCircuits]Circuit|
|NumColors         |uint32              |
|Colors            |[NumColors]Color    |
|NumPumps          |uint32              |
|Pumps             |[NumPumps]Pump      |
|InterfaceTabFlags |uint32              |
|ShowAlarms        |bool                |

**Circuit**

|Field        |Type  |
|-------------|------|
|ID           |uint32|
|Name         |string|
|NameIndex    |uint8 |
|Function     |uint8 |
|Interface    |uint8 |
|Flags        |uint8 |
|ColorSet     |uint8 |
|ColorPosition|uint8 |
|ColorStagger |uint8 |
|DeviceID     |uint8 |
|DefaultRT    |uint16|

**Color**

|Field|Type  |
|-----|------|
|Name |String|
|Red  |uint32|
|Green|uint32|
|Blue |uint32|

**Pump**

|Field|Type |
|-----|-----|
|Data |uint8|

## Error Packet Types

### Login Failed

This packet consists of a header only, who's `Code` field is `13`.

### Bad Parameter

This packet consists of a header only, who's `Code` field is `31`.