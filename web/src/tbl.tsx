import { 
    List, 
    Edit, 
    Show,
    Create,
    Datagrid, 
    TextField, 
    ReferenceField, 
    ReferenceInput, 
    SimpleForm, 
    TextInput,
    SelectInput,
    SelectField,
    SimpleShowLayout
} from 'react-admin';

import { TableTypeInput } from './common';

export const TblList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" label="ID"/>
            <TextField source="db_name" label = "Database"/>
            <TextField source="schema" />
            <TextField source="name" />
            <TextField source="type" />
            <TextField source="target" />
            <TextField source="partitions_regex" />
        </Datagrid>
    </List>
);

export const TblEdit = () => (
    <Edit>
        <SimpleForm>
            <TextField source="id" label="ID"/>
            <TextField source="db_name" label = "Database"/>
            <TextInput source="schema" fullWidth />
            <TextInput source="name" fullWidth />
            <TableTypeInput />
            <TextInput source="target" fullWidth />
            <TextInput source="partitions_regex" fullWidth />
        </SimpleForm>
    </Edit>
);

export const TblShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" label="ID"/>
            <TextField source="db_name" label = "Database"/>
            <TextField source="schema" />
            <TextField source="name" />
            <TextField source="type" />
            <TextField source="target" />
            <TextField source="partitions_regex" />
        </SimpleShowLayout>
    </Show>
);

export const TblCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <ReferenceInput source="db_id" reference="db" label="DB" >
                <SelectInput optionText="name"/>
            </ReferenceInput>
            <TextInput source="schema" fullWidth />
            <TextInput source="name" fullWidth />
            <TableTypeInput />
            <TextInput source="target" fullWidth />
            <TextInput source="partitions_regex" fullWidth />
        </SimpleForm>
    </Create>
);
