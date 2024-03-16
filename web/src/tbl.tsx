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
    SimpleShowLayout
} from 'react-admin';

import { TableTypeInput } from './common';

export const TblList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" label="ID"/>
            <ReferenceField source="DBId" reference="db" label="DB" />
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
            <ReferenceField source="DBId" reference="db" label="DB" />
            <TextInput source="name" />
            <TableTypeInput />
            <TextInput source="target" />
            <TextInput source="partitions_regex" />
        </SimpleForm>
    </Edit>
);

export const TblShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" label="ID"/>
            <ReferenceField source="DBId" reference="db" label="DB" />
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
            <ReferenceInput source="DBId" reference="db" label="DB" />
            <TextInput source="name" />
            <TableTypeInput />
            <TextInput source="target" />
            <TextInput source="partitions_regex" />
        </SimpleForm>
    </Create>
);
