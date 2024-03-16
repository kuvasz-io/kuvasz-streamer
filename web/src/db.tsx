import { 
    Datagrid, 
    List, 
    Edit, 
    Show,
    Create,
    TextField, 
    ReferenceField, 
    TextInput,
    ReferenceInput, 
    SimpleForm, 
    SimpleShowLayout
} from 'react-admin';


export const DbList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" label="ID" />
            <TextField source="name" />
        </Datagrid>
    </List>
);

export const DbEdit = () => (
    <Edit>
        <SimpleForm>
            <TextField source="id" label="ID" />
            <TextInput source="name" />
        </SimpleForm>
    </Edit>
);

export const DbShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" label="ID"  />
            <TextField source="name" />
        </SimpleShowLayout>
    </Show>
);

export const DbCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <TextInput source="name" multiline={false} label="Name" />
        </SimpleForm>
    </Create>
);
