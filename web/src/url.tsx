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

export const UrlList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" label="ID"/>
            <ReferenceField source="db_id" reference="db"  label="DB"/>
            <TextField source="sid" label="SID"/>
            <TextField source="url" label="URL"/>
        </Datagrid>
    </List>
);

export const UrlEdit = () => (
    <Edit>
        <SimpleForm>
            <TextField source="id" label="ID"/>
            <ReferenceField source="db_id" reference="db" label="DB"/>
            <TextInput source="sid" label="SID"/>
            <TextInput source="url" label="URL" fullWidth />
        </SimpleForm>
    </Edit>
);

export const UrlShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" label="ID"/>
            <ReferenceField source="db_id" reference="db" label="DB"/>
            <TextField source="sid" label="SID"/>
            <TextField source="url" label="URL" />
        </SimpleShowLayout>
    </Show>
);

export const UrlCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <ReferenceInput source="db_id" reference="db" label="DB"/>
            <TextInput source="sid" label="SID"/>
            <TextInput source="url" label="URL" fullWidth/>
        </SimpleForm>
    </Create>
);
