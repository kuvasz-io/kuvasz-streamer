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

export const DbList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" />
            <TextField source="name" />
        </Datagrid>
    </List>
);

export const DbEdit = () => (
    <Edit>
        <SimpleForm>
            <TextInput source="id" />
            <TextInput source="name" />
        </SimpleForm>
    </Edit>
);

export const DbShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextInput source="id" />
            <TextInput source="name" />
        </SimpleShowLayout>
    </Show>
);

