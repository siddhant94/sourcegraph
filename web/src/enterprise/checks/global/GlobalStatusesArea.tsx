import { StatusScope } from '@sourcegraph/extension-api-classes'
import React from 'react'
import { RouteComponentProps } from 'react-router'
import { ChecksArea, ChecksAreaContext } from '../../status/statusesArea/ChecksArea'

interface Props
    extends Pick<ChecksAreaContext, Exclude<keyof ChecksAreaContext, 'scope'>>,
        RouteComponentProps<{}> {}

/**
 * The global statuses area.
 */
export const GlobalChecksArea: React.FunctionComponent<Props> = ({ ...props }) => (
    <ChecksArea {...props} scope={StatusScope.Global} />
)
