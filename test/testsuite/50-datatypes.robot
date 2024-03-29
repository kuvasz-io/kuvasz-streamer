*** Settings ***
Resource           00-common.robot
Suite Setup        Connect To All Databases
Suite Teardown     Disconnect From All Databases
Test Template      Update field should propagate

*** Keywords ***
Update field should propagate
    [Arguments]    ${FIELD}    ${VALUE}
    Statement should propagate
    ...    update t5 set ${FIELD} = ${VALUE}
    ...    Select ${FIELD} from t5
    ...    Select ${FIELD} from t5 where sid='{}' 

*** Test cases ***
bigint             f1    -9023372036854770000
bigserial          f2    1
bit                f3    '1'
bit varying        f4    '110101'
boolean            f5    true
box                f6    '(1,1),(4,4)'
bytea              f7    '\xdeadbeef'
character          f8    'A'
varchar            f9    'ABCD'
cidr               f10    '192.168.0.1'
circle             f11    '(2,2),4'
date               f12    '2023-01-01'
double precision   f13    123.123456789012345
inet               f14    '192.168.0.0/16'
integer            f15    1000000
interval           f16    'P1DT5M'
json               f17    '{"name":"value"}'
jsonb              f18    '{"name":"value"}'
line               f19    '{1,2,3}'
lseg               f20    '(1,1),(5,5)'
macaddr            f21    '08:00:2b:01:02:03'
macaddr8           f22    '08:00:2b:01:02:03:04:05'
money              f23    123.12
numeric            f24    1234567890.12345678901234567890
path               f25    '[(1,1),(2,1),(4,4)]'
pg_lsn             f26    '16/B374D848'
point              f28    '(1,2)'
polygon            f29    '(1,1),(2,1),(4,4)'
real               f30    123.123456
smallint           f31    32000
smallserial        f32    1
serial             f33    1
text               f34    'abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij'
time               f35    '12:34:56.123'
timetz             f36    '12:34:56.123+02'
timestamp          f37    '2023-01-02 01:02:03.123'
timestamptz        f38    '2023-01-02 01:02:03.123+02'
tsquery            f39    'fat & rat'
tsvector           f40    'a fat cat sat on a mat and ate a fat rat'
uuid               f42    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
xml                f43    '<foo>bar</foo>'
integer[]          f44    '{1,2,3,4}'
complex            f45    '(1,2)'
