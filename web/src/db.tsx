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
    SimpleShowLayout,
    BulkDeleteButton
} from 'react-admin';


export const DbList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" label="ID" />
            <TextField source="name" label="Database"/>
        </Datagrid>
    </List>
);

export const DbView = () => (
    <List>
        <Datagrid bulkActionButtons={false}>
            <TextField source="id" label="ID" />
            <TextField source="name" label="Database"/>
        </Datagrid>
    </List>
);

export const DbEdit = () => (
    <Edit>
        <SimpleForm>
            <TextField source="id" label="ID" />
            <TextInput source="name" label="Database"/>
        </SimpleForm>
    </Edit>
);

export const DbShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" label="ID"  />
            <TextField source="name" label="Database"/>
        </SimpleShowLayout>
    </Show>
);

export const DbCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <TextInput source="name" multiline={false} label="Database"/>
        </SimpleForm>
    </Create>
);
