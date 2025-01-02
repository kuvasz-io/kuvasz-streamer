*** Settings ***
Resource           00-common.robot
Suite Setup        Setup empty record
Suite Teardown     Disconnect From All Databases

*** Variables ***
${FIELDS}          f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f14,f15,f16,f17,f18,f19,f20,f21,f22,f23,f24,f25,f26,f28,f29,f30,f31,f32,f33,f34,f35,f36,f37,f38,f39,f40,f42,f44,f45
${PAIRS}          SEPARATOR=\n
...                f1=-9023372036854770000,
...                f2=1,
...                f3='1',
...                f4 ='110101',
...                f5 =true,
...                f6 ='(1,1),(4,4)',
...                f7 ='\xdeadbeef',
...                f8 ='A',
...                f9 ='ABCD',
...                f10= '192.168.0.1',
...                f11= '(2,2),4',
...                f12= '2023-01-01',
...                f13= 123.123456789012345,
...                f14= '192.168.0.0/16',
...                f15= 1000000,
...                f16= 'P1DT5M',
...                f17= '{"name":"value"}',
...                f18= '{"name":"value"}',
...                f19= '{1,2,3}',
...                f20= '(1,1),(5,5)',
...                f21= '08:00:2b:01:02:03',
...                f22= '08:00:2b:01:02:03:04:05',
...                f23= 123.12,
...                f24= 1234567890.12345678901234567890,
...                f25= '[(1,1),(2,1),(4,4)]',
...                f26= '16/B374D848',
...                f28= '(1,2)',
...                f29= '(1,1),(2,1),(4,4)',
...                f30= 123.123456,
...                f31= 32000,
...                f32= 1,
...                f33= 1,
...                f34= 'abcdefghijabcdefghijabcdefghijabcdefghijabcdefghij',
...                f35= '12:34:56.123',
...                f36= '12:34:56.123+02',
...                f37= '2023-01-02 01:02:03.123',
...                f38= '2023-01-02 01:02:03.123+02',
...                f39= 'fat & rat',
...                f40= 'a fat cat sat on a mat and ate a fat rat',
...                f42= 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
...                f44= '{1,2,3,4}',
...                f45= '(1,2)'

*** Keywords ***
Setup empty record
    Connect To All Databases
    FOR    ${PG}    IN    @{PGVERSIONS}
        Switch database       ${PG}
        Execute SQL string    insert into t5(f1) values(null)
    END

*** Test cases ***
Update all fields should propagate
    Statement should propagate
    ...    update t5 set ${PAIRS}
    ...    Select ${FIELDS} from t5
    ...    Select ${FIELDS} from t5 where sid='{}'

