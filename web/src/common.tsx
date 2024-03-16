import * as React from 'react';
import {
    SelectInput,
    SaveButton,
    Toolbar,
    useNotify,
} from 'react-admin';
import { useFormContext } from 'react-hook-form';

export const TableTypeInput = () => (
    <SelectInput source="type" choices={[
        { id: 'clone', name: 'clone'},
        { id: 'append', name: 'append' },
        { id: 'history', name: 'history' },
    ]} />
)


export const PostCreateToolbar = () => {
    const notify = useNotify();
    const { reset } = useFormContext();

    return (
        <Toolbar>
            <SaveButton
                type="button"
                label="post.action.save_and_add"
                variant="text"
                mutationOptions={{
                    onSuccess: () => {
                        reset();
                        window.scrollTo(0, 0);
                        notify('ra.notification.created', {
                            type: 'info',
                            messageArgs: { smart_count: 1 },
                        });
                    },
                }}
            />
        </Toolbar>
    );
};
