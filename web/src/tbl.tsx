import { 
    Datagrid, 
    List, 
    TextField, 
    Edit, 
    ReferenceField, 
    ReferenceInput, 
    SimpleForm, 
    TextInput,
    Show,
    SimpleShowLayout
} from 'react-admin';

export const TblList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" />
            <ReferenceField source="DBId" reference="db" />
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
            <TextInput source="id" />
            <ReferenceInput source="DBId" reference="db" />
            <TextInput source="name" />
            <TextInput source="type" />
            <TextInput source="target" />
            <TextInput source="partitions_regex" />
        </SimpleForm>
    </Edit>
);

export const TblShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" />
            <ReferenceField source="DBId" reference="db" />
            <TextField source="name" />
            <TextField source="type" />
            <TextField source="target" />
            <TextField source="partitions_regex" />
        </SimpleShowLayout>
    </Show>
);
