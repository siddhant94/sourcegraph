import classNames from 'classnames'
import * as React from 'react'

import styles from './BlobStatusBarContainer.module.scss'

interface BlobStatusBarContainerProps {
    className?: string
}

export const BlobStatusBarContainer: React.FunctionComponent<BlobStatusBarContainerProps> = ({
    children,
    className,
}) => <div className={classNames(className, styles.blobStatusBarContainer)}>{children}</div>
