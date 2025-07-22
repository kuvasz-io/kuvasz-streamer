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
    SimpleShowLayout,
    BooleanField,
    Button,
    useDataProvider, 
    useRecordContext, 
    useRefresh,
    useDelete, 
    useNotify
} from 'react-admin';

import { useMutation } from '@tanstack/react-query';

import { TableTypeInput } from './common';

const CloneButton = () => {
    const record = useRecordContext();
    const dataProvider = useDataProvider();
    const refresh = useRefresh();

    const { mutate, isLoading } = useMutation(
        () => dataProvider.cloneTable(record.id).then(() => refresh()));
    if (!record.replicated) {
        return <Button 
                label="Clone" 
                onClick={() => mutate()}
                disabled={isLoading} />;
    }
    return null;
};

const RefreshButton = () => {
    const refresh = useRefresh();
    const handleClick = () => {
        refresh();
    }
    return <button onClick={handleClick}>Refresh</button>;
};

export const MapList = () => (
    <List pagination={false} perPage={1000}>
        <Datagrid bulkActionButtons={false}>
            <TextField source="id" sortable={false}/>
            <TextField source="db_id" label ="DB ID"sortable={false}/>
            <TextField source="db_name" label="Database" sortable={false}/>
            <TextField source="schema" label="Schema" sortable={false}/>
            <TextField source="name" label="Table" sortable={false}/>
            <TextField source="type" sortable={false}/>
            <TextField source="target" sortable={false}/>
            <TextField source="partitions" sortable={false}/>
            <TextField source="partitions_regex" sortable={false}/>
            <BooleanField    source="replicated" sortable={false}/>
            <BooleanField    source="present" sortable={false}/>
            <CloneButton />
        </Datagrid>
    </List>
);

export const MapEdit = () => (
    <Edit>
        <Datagrid>
        <TextField source="id" />
            <TextField source="db_id" />
            <TextField source="name" />
            <TextField source="schema" />
            <TextField source="type" />
            <TextField source="target" />
            <TextField source="partitions_regex" />
            <BooleanField    source="replicated" />
            <BooleanField    source="present" />
        </Datagrid>
    </Edit>
);






