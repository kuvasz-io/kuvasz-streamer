import { 
    Identifier,
    RaRecord,
    List, 
    Edit, 
    Show,
    Create,
    Datagrid, 
    TextField, 
    ReferenceField, 
    BooleanField,
    ReferenceInput, 
    SimpleForm, 
    TextInput,
    SelectInput,
    SimpleShowLayout,
    EditButton,
    RecordContext
} from 'react-admin';

import { useRecordContext } from 'react-admin';
import { Button, Link } from '@mui/material';  

import WarningIcon from '@mui/icons-material/Warning';

export const UrlList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" label="ID"/>
            <ReferenceField source="db_id" reference="db"  label="Database">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="sid" label="SID"/>
            <TextField source="url" label="URL"/>
            <BooleanField source='up' label="Status" FalseIcon={WarningIcon}/>
        </Datagrid>
    </List>
);

export const UrlEdit = () => (
    <Edit>
        <SimpleForm>
            <TextField source="id" label="ID"/>
            <ReferenceField source="db_id" reference="db"  label="Database">
                <TextField source="name" />
            </ReferenceField>
            <TextInput source="sid" label="SID"/>
            <TextInput source="url" label="URL" fullWidth />
        </SimpleForm>
    </Edit>
);

export const UrlShow = () => (
    <Show>
        <SimpleShowLayout>
            <TextField source="id" label="ID"/>
            <ReferenceField source="db_id" reference="db"  label="Database">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="sid" label="SID"/>
            <TextField source="url" label="URL" />
        </SimpleShowLayout>
    </Show>
);

export const UrlCreate = () => (
    <Create redirect="list">
        <SimpleForm>
            <ReferenceInput source="db_id" reference="db" label="Database">
                <SelectInput optionText="name"/>
            </ReferenceInput>
            <TextInput source="sid" label="SID"/>
            <TextInput source="url" label="URL" fullWidth/>
        </SimpleForm>
    </Create>
);
