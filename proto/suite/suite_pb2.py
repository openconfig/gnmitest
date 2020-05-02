# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: proto/suite/suite.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from github.com.openconfig.gnmitest.proto.tests import tests_pb2 as github_dot_com_dot_openconfig_dot_gnmitest_dot_proto_dot_tests_dot_tests__pb2
from github.com.openconfig.gnmi.proto.gnmi import gnmi_pb2 as github_dot_com_dot_openconfig_dot_gnmi_dot_proto_dot_gnmi_dot_gnmi__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='proto/suite/suite.proto',
  package='suite',
  syntax='proto3',
  serialized_pb=_b('\n\x17proto/suite/suite.proto\x12\x05suite\x1a\x36github.com/openconfig/gnmitest/proto/tests/tests.proto\x1a\x30github.com/openconfig/gnmi/proto/gnmi/gnmi.proto\"\xbc\x02\n\x05Suite\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0f\n\x07timeout\x18\x02 \x01(\x05\x12\x0e\n\x06schema\x18\x03 \x01(\t\x12%\n\nconnection\x18\x04 \x01(\x0b\x32\x11.tests.Connection\x12\x37\n\x0e\x65xtension_list\x18\n \x03(\x0b\x32\x1f.suite.Suite.ExtensionListEntry\x12\x31\n\x13instance_group_list\x18\x0f \x03(\x0b\x32\x14.suite.InstanceGroup\x12%\n\x06\x63ommon\x18\x10 \x01(\x0b\x32\x15.suite.CommonMessages\x1aJ\n\x12\x45xtensionListEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12#\n\x05value\x18\x02 \x01(\x0b\x32\x14.suite.ExtensionList:\x02\x38\x01\"V\n\rInstanceGroup\x12\x13\n\x0b\x64\x65scription\x18\x01 \x01(\t\x12!\n\x08instance\x18\x02 \x03(\x0b\x32\x0f.suite.Instance\x12\r\n\x05\x66\x61tal\x18\x03 \x01(\x08\"R\n\x08Instance\x12\x13\n\x0b\x64\x65scription\x18\x01 \x01(\t\x12\x19\n\x04test\x18\x02 \x01(\x0b\x32\x0b.tests.Test\x12\x16\n\x0e\x65xtension_list\x18\x03 \x03(\t\"/\n\rExtensionList\x12\x1e\n\textension\x18\x01 \x03(\x0b\x32\x0b.tests.Test\"\xbc\x04\n\x0e\x43ommonMessages\x12<\n\x0cset_requests\x18\x01 \x03(\x0b\x32&.suite.CommonMessages.SetRequestsEntry\x12<\n\x0cget_requests\x18\x03 \x03(\x0b\x32&.suite.CommonMessages.GetRequestsEntry\x12>\n\rget_responses\x18\x04 \x03(\x0b\x32\'.suite.CommonMessages.GetResponsesEntry\x12H\n\x12subscribe_requests\x18\x05 \x03(\x0b\x32,.suite.CommonMessages.SubscribeRequestsEntry\x1a\x44\n\x10SetRequestsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\x1f\n\x05value\x18\x02 \x01(\x0b\x32\x10.gnmi.SetRequest:\x02\x38\x01\x1a\x44\n\x10GetRequestsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\x1f\n\x05value\x18\x02 \x01(\x0b\x32\x10.gnmi.GetRequest:\x02\x38\x01\x1a\x46\n\x11GetResponsesEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12 \n\x05value\x18\x02 \x01(\x0b\x32\x11.gnmi.GetResponse:\x02\x38\x01\x1aP\n\x16SubscribeRequestsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12%\n\x05value\x18\x02 \x01(\x0b\x32\x16.gnmi.SubscribeRequest:\x02\x38\x01\x62\x06proto3')
  ,
  dependencies=[github_dot_com_dot_openconfig_dot_gnmitest_dot_proto_dot_tests_dot_tests__pb2.DESCRIPTOR,github_dot_com_dot_openconfig_dot_gnmi_dot_proto_dot_gnmi_dot_gnmi__pb2.DESCRIPTOR,])




_SUITE_EXTENSIONLISTENTRY = _descriptor.Descriptor(
  name='ExtensionListEntry',
  full_name='suite.Suite.ExtensionListEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='suite.Suite.ExtensionListEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='suite.Suite.ExtensionListEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=_descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001')),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=383,
  serialized_end=457,
)

_SUITE = _descriptor.Descriptor(
  name='Suite',
  full_name='suite.Suite',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='suite.Suite.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='timeout', full_name='suite.Suite.timeout', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='schema', full_name='suite.Suite.schema', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='connection', full_name='suite.Suite.connection', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='extension_list', full_name='suite.Suite.extension_list', index=4,
      number=10, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='instance_group_list', full_name='suite.Suite.instance_group_list', index=5,
      number=15, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='common', full_name='suite.Suite.common', index=6,
      number=16, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[_SUITE_EXTENSIONLISTENTRY, ],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=141,
  serialized_end=457,
)


_INSTANCEGROUP = _descriptor.Descriptor(
  name='InstanceGroup',
  full_name='suite.InstanceGroup',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='description', full_name='suite.InstanceGroup.description', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='instance', full_name='suite.InstanceGroup.instance', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='fatal', full_name='suite.InstanceGroup.fatal', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=459,
  serialized_end=545,
)


_INSTANCE = _descriptor.Descriptor(
  name='Instance',
  full_name='suite.Instance',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='description', full_name='suite.Instance.description', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='test', full_name='suite.Instance.test', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='extension_list', full_name='suite.Instance.extension_list', index=2,
      number=3, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=547,
  serialized_end=629,
)


_EXTENSIONLIST = _descriptor.Descriptor(
  name='ExtensionList',
  full_name='suite.ExtensionList',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='extension', full_name='suite.ExtensionList.extension', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=631,
  serialized_end=678,
)


_COMMONMESSAGES_SETREQUESTSENTRY = _descriptor.Descriptor(
  name='SetRequestsEntry',
  full_name='suite.CommonMessages.SetRequestsEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='suite.CommonMessages.SetRequestsEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='suite.CommonMessages.SetRequestsEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=_descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001')),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=961,
  serialized_end=1029,
)

_COMMONMESSAGES_GETREQUESTSENTRY = _descriptor.Descriptor(
  name='GetRequestsEntry',
  full_name='suite.CommonMessages.GetRequestsEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='suite.CommonMessages.GetRequestsEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='suite.CommonMessages.GetRequestsEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=_descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001')),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1031,
  serialized_end=1099,
)

_COMMONMESSAGES_GETRESPONSESENTRY = _descriptor.Descriptor(
  name='GetResponsesEntry',
  full_name='suite.CommonMessages.GetResponsesEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='suite.CommonMessages.GetResponsesEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='suite.CommonMessages.GetResponsesEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=_descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001')),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1101,
  serialized_end=1171,
)

_COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY = _descriptor.Descriptor(
  name='SubscribeRequestsEntry',
  full_name='suite.CommonMessages.SubscribeRequestsEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='suite.CommonMessages.SubscribeRequestsEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='suite.CommonMessages.SubscribeRequestsEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=_descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001')),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1173,
  serialized_end=1253,
)

_COMMONMESSAGES = _descriptor.Descriptor(
  name='CommonMessages',
  full_name='suite.CommonMessages',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='set_requests', full_name='suite.CommonMessages.set_requests', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='get_requests', full_name='suite.CommonMessages.get_requests', index=1,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='get_responses', full_name='suite.CommonMessages.get_responses', index=2,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='subscribe_requests', full_name='suite.CommonMessages.subscribe_requests', index=3,
      number=5, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[_COMMONMESSAGES_SETREQUESTSENTRY, _COMMONMESSAGES_GETREQUESTSENTRY, _COMMONMESSAGES_GETRESPONSESENTRY, _COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY, ],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=681,
  serialized_end=1253,
)

_SUITE_EXTENSIONLISTENTRY.fields_by_name['value'].message_type = _EXTENSIONLIST
_SUITE_EXTENSIONLISTENTRY.containing_type = _SUITE
_SUITE.fields_by_name['connection'].message_type = github_dot_com_dot_openconfig_dot_gnmitest_dot_proto_dot_tests_dot_tests__pb2._CONNECTION
_SUITE.fields_by_name['extension_list'].message_type = _SUITE_EXTENSIONLISTENTRY
_SUITE.fields_by_name['instance_group_list'].message_type = _INSTANCEGROUP
_SUITE.fields_by_name['common'].message_type = _COMMONMESSAGES
_INSTANCEGROUP.fields_by_name['instance'].message_type = _INSTANCE
_INSTANCE.fields_by_name['test'].message_type = github_dot_com_dot_openconfig_dot_gnmitest_dot_proto_dot_tests_dot_tests__pb2._TEST
_EXTENSIONLIST.fields_by_name['extension'].message_type = github_dot_com_dot_openconfig_dot_gnmitest_dot_proto_dot_tests_dot_tests__pb2._TEST
_COMMONMESSAGES_SETREQUESTSENTRY.fields_by_name['value'].message_type = github_dot_com_dot_openconfig_dot_gnmi_dot_proto_dot_gnmi_dot_gnmi__pb2._SETREQUEST
_COMMONMESSAGES_SETREQUESTSENTRY.containing_type = _COMMONMESSAGES
_COMMONMESSAGES_GETREQUESTSENTRY.fields_by_name['value'].message_type = github_dot_com_dot_openconfig_dot_gnmi_dot_proto_dot_gnmi_dot_gnmi__pb2._GETREQUEST
_COMMONMESSAGES_GETREQUESTSENTRY.containing_type = _COMMONMESSAGES
_COMMONMESSAGES_GETRESPONSESENTRY.fields_by_name['value'].message_type = github_dot_com_dot_openconfig_dot_gnmi_dot_proto_dot_gnmi_dot_gnmi__pb2._GETRESPONSE
_COMMONMESSAGES_GETRESPONSESENTRY.containing_type = _COMMONMESSAGES
_COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY.fields_by_name['value'].message_type = github_dot_com_dot_openconfig_dot_gnmi_dot_proto_dot_gnmi_dot_gnmi__pb2._SUBSCRIBEREQUEST
_COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY.containing_type = _COMMONMESSAGES
_COMMONMESSAGES.fields_by_name['set_requests'].message_type = _COMMONMESSAGES_SETREQUESTSENTRY
_COMMONMESSAGES.fields_by_name['get_requests'].message_type = _COMMONMESSAGES_GETREQUESTSENTRY
_COMMONMESSAGES.fields_by_name['get_responses'].message_type = _COMMONMESSAGES_GETRESPONSESENTRY
_COMMONMESSAGES.fields_by_name['subscribe_requests'].message_type = _COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY
DESCRIPTOR.message_types_by_name['Suite'] = _SUITE
DESCRIPTOR.message_types_by_name['InstanceGroup'] = _INSTANCEGROUP
DESCRIPTOR.message_types_by_name['Instance'] = _INSTANCE
DESCRIPTOR.message_types_by_name['ExtensionList'] = _EXTENSIONLIST
DESCRIPTOR.message_types_by_name['CommonMessages'] = _COMMONMESSAGES
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Suite = _reflection.GeneratedProtocolMessageType('Suite', (_message.Message,), dict(

  ExtensionListEntry = _reflection.GeneratedProtocolMessageType('ExtensionListEntry', (_message.Message,), dict(
    DESCRIPTOR = _SUITE_EXTENSIONLISTENTRY,
    __module__ = 'proto.suite.suite_pb2'
    # @@protoc_insertion_point(class_scope:suite.Suite.ExtensionListEntry)
    ))
  ,
  DESCRIPTOR = _SUITE,
  __module__ = 'proto.suite.suite_pb2'
  # @@protoc_insertion_point(class_scope:suite.Suite)
  ))
_sym_db.RegisterMessage(Suite)
_sym_db.RegisterMessage(Suite.ExtensionListEntry)

InstanceGroup = _reflection.GeneratedProtocolMessageType('InstanceGroup', (_message.Message,), dict(
  DESCRIPTOR = _INSTANCEGROUP,
  __module__ = 'proto.suite.suite_pb2'
  # @@protoc_insertion_point(class_scope:suite.InstanceGroup)
  ))
_sym_db.RegisterMessage(InstanceGroup)

Instance = _reflection.GeneratedProtocolMessageType('Instance', (_message.Message,), dict(
  DESCRIPTOR = _INSTANCE,
  __module__ = 'proto.suite.suite_pb2'
  # @@protoc_insertion_point(class_scope:suite.Instance)
  ))
_sym_db.RegisterMessage(Instance)

ExtensionList = _reflection.GeneratedProtocolMessageType('ExtensionList', (_message.Message,), dict(
  DESCRIPTOR = _EXTENSIONLIST,
  __module__ = 'proto.suite.suite_pb2'
  # @@protoc_insertion_point(class_scope:suite.ExtensionList)
  ))
_sym_db.RegisterMessage(ExtensionList)

CommonMessages = _reflection.GeneratedProtocolMessageType('CommonMessages', (_message.Message,), dict(

  SetRequestsEntry = _reflection.GeneratedProtocolMessageType('SetRequestsEntry', (_message.Message,), dict(
    DESCRIPTOR = _COMMONMESSAGES_SETREQUESTSENTRY,
    __module__ = 'proto.suite.suite_pb2'
    # @@protoc_insertion_point(class_scope:suite.CommonMessages.SetRequestsEntry)
    ))
  ,

  GetRequestsEntry = _reflection.GeneratedProtocolMessageType('GetRequestsEntry', (_message.Message,), dict(
    DESCRIPTOR = _COMMONMESSAGES_GETREQUESTSENTRY,
    __module__ = 'proto.suite.suite_pb2'
    # @@protoc_insertion_point(class_scope:suite.CommonMessages.GetRequestsEntry)
    ))
  ,

  GetResponsesEntry = _reflection.GeneratedProtocolMessageType('GetResponsesEntry', (_message.Message,), dict(
    DESCRIPTOR = _COMMONMESSAGES_GETRESPONSESENTRY,
    __module__ = 'proto.suite.suite_pb2'
    # @@protoc_insertion_point(class_scope:suite.CommonMessages.GetResponsesEntry)
    ))
  ,

  SubscribeRequestsEntry = _reflection.GeneratedProtocolMessageType('SubscribeRequestsEntry', (_message.Message,), dict(
    DESCRIPTOR = _COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY,
    __module__ = 'proto.suite.suite_pb2'
    # @@protoc_insertion_point(class_scope:suite.CommonMessages.SubscribeRequestsEntry)
    ))
  ,
  DESCRIPTOR = _COMMONMESSAGES,
  __module__ = 'proto.suite.suite_pb2'
  # @@protoc_insertion_point(class_scope:suite.CommonMessages)
  ))
_sym_db.RegisterMessage(CommonMessages)
_sym_db.RegisterMessage(CommonMessages.SetRequestsEntry)
_sym_db.RegisterMessage(CommonMessages.GetRequestsEntry)
_sym_db.RegisterMessage(CommonMessages.GetResponsesEntry)
_sym_db.RegisterMessage(CommonMessages.SubscribeRequestsEntry)


_SUITE_EXTENSIONLISTENTRY.has_options = True
_SUITE_EXTENSIONLISTENTRY._options = _descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001'))
_COMMONMESSAGES_SETREQUESTSENTRY.has_options = True
_COMMONMESSAGES_SETREQUESTSENTRY._options = _descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001'))
_COMMONMESSAGES_GETREQUESTSENTRY.has_options = True
_COMMONMESSAGES_GETREQUESTSENTRY._options = _descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001'))
_COMMONMESSAGES_GETRESPONSESENTRY.has_options = True
_COMMONMESSAGES_GETRESPONSESENTRY._options = _descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001'))
_COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY.has_options = True
_COMMONMESSAGES_SUBSCRIBEREQUESTSENTRY._options = _descriptor._ParseOptions(descriptor_pb2.MessageOptions(), _b('8\001'))
# @@protoc_insertion_point(module_scope)
