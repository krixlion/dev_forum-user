# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [user_service.proto](#user_service-proto)
    - [CreateUserRequest](#user-CreateUserRequest)
    - [CreateUserResponse](#user-CreateUserResponse)
    - [DeleteUserRequest](#user-DeleteUserRequest)
    - [GetUserRequest](#user-GetUserRequest)
    - [GetUserResponse](#user-GetUserResponse)
    - [GetUserSecretRequest](#user-GetUserSecretRequest)
    - [GetUserSecretResponse](#user-GetUserSecretResponse)
    - [GetUsersRequest](#user-GetUsersRequest)
    - [UpdateUserRequest](#user-UpdateUserRequest)
    - [User](#user-User)
  
    - [UserService](#user-UserService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="user_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## user_service.proto



<a name="user-CreateUserRequest"></a>

### CreateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#user-User) |  |  |






<a name="user-CreateUserResponse"></a>

### CreateUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="user-DeleteUserRequest"></a>

### DeleteUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="user-GetUserRequest"></a>

### GetUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="user-GetUserResponse"></a>

### GetUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#user-User) |  |  |






<a name="user-GetUserSecretRequest"></a>

### GetUserSecretRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| email | [string](#string) |  |  |






<a name="user-GetUserSecretResponse"></a>

### GetUserSecretResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#user-User) |  |  |






<a name="user-GetUsersRequest"></a>

### GetUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| offset | [string](#string) |  |  |
| limit | [string](#string) |  |  |
| filter | [string](#string) |  |  |






<a name="user-UpdateUserRequest"></a>

### UpdateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#user-User) |  |  |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  |  |






<a name="user-User"></a>

### User



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |
| password | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |





 

 

 


<a name="user-UserService"></a>

### UserService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateUserRequest](#user-CreateUserRequest) | [CreateUserResponse](#user-CreateUserResponse) |  |
| Update | [UpdateUserRequest](#user-UpdateUserRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| Delete | [DeleteUserRequest](#user-DeleteUserRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| Get | [GetUserRequest](#user-GetUserRequest) | [GetUserResponse](#user-GetUserResponse) |  |
| GetSecret | [GetUserSecretRequest](#user-GetUserSecretRequest) | [GetUserSecretResponse](#user-GetUserSecretResponse) |  |
| GetStream | [GetUsersRequest](#user-GetUsersRequest) | [User](#user-User) stream |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

