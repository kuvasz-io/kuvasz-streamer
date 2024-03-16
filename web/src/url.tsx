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

export const UrlList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" />
            <ReferenceField source="db_id" reference="db" />
            <TextField source="sid" />
            <TextField source="url" />
        </Datagrid>
    </List>
);

export const UrlEdit = () => (
    <Edit>
        <SimpleForm>
            <TextInput source="id" />
            <ReferenceInput source="db_id" reference="db" />
            <TextInput source="sid" />
            <TextInput source="url" />
        </SimpleForm>
    </Edit>
);

export const UrlShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" />
            <ReferenceField source="db_id" reference="db" />
            <TextField source="sid" />
            <TextField source="url" />
        </SimpleShowLayout>
    </Show>
);

