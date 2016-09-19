// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import * as layout from "sourcegraph/components/styles/_layout.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as styles from "sourcegraph/desktop/styles/home.css";
import * as classNames from "classnames";

import {Link} from "react-router";
import {Heading, List} from "sourcegraph/components";
import {Cone} from "sourcegraph/components/symbols";

export const NotInBeta = () => (
	<div className={classNames(layout.containerFixed, base.pv5, base.ph4)} style={{maxWidth: "600px"}}>
	<Heading align="center" level="4" underline="blue">
		It looks like you're not in the desktop beta right now.
	</Heading>
	</div>
);

export class DesktopHome extends React.Component<{}, any> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		return (
			<div className={classNames(layout.containerFixed, base.pv5, base.ph4)} style={{maxWidth: "560px"}}>
				<Heading align="center" level="4" underline="blue">
					See live examples, search code, and view inline
					<br className={base.hidden_s} />&nbsp;documentation to write better code, faster
				</Heading>

				<img src={`${(this.context as any).siteConfig.assetsRoot}/img/sg-desktop.gif`} width="356" title="Usage examples right in your editor" alt="Usage examples right in your editor" style={{maxWidth: "100%", display: "block", imageRendering: "pixelated"}} className={base.center}/>

				<div className={base.mv4}>
					<Heading level="5">Go definitions and usages as you code</Heading>
					<p>
						Install one of our <a href="/integrations">editor integrations,</a> and as you write Go code, this pane will update with contextually relevant information.
					</p>
				</div>
				<div className={base.mv4}>
					<Heading level="5">Semantic, global code search</Heading>
					<p>
						Just hit <span className={styles.label_blue}>⌘ or CTRL </span> + <span className={styles.label_blue}>SHIFT</span> + <span className={styles.label_blue}>;</span> or click the search box at the top of this page to semantically search for functions and symbols.
					</p>
				</div>
				<div className={base.mv4}>
					<Heading level="5">Powerful search for your private code</Heading>
					<p>
						To enable semantic search and usage examples for your private code, <Link to="/settings">authorize Sourcegraph</Link> to access your private repositories.
					</p>
				</div>
				<div className={classNames(base.mt5, typography.f7)}>
					<Heading level="6">
						<Cone width={16} className={classNames(colors.fill_orange, base.mr2)} style={{
							verticalAlign: "baseline",
							position: "relative",
							top: "1px",
						}} />
						Sourcegraph Desktop is currently in beta
					</Heading>
					<p>
						Thanks for using Sourcegraph Desktop! If the app is not working as expected, see our GitHub to:
					</p>
					<List className={base.mv3}>
						<li><strong><a target="_blank" href="https://github.com/sourcegraph-beta/sourcegraph-desktop/blob/master/troubleshooting.md#sourcegraph-desktop-troubleshooting">Browse troubleshooting tips</a></strong></li>
						<li><strong><a target="_blank" href="https://github.com/sourcegraph-beta/sourcegraph-desktop/issues/new">File an issue</a></strong></li>
					</List>
					<p>
						We love feedback! Shoot us an email at <strong><a href="mailto:support@sourcegraph.com?subject=Feedback for the Sourcegraph Desktop team&body=Editor of choice: %0D%0A%0D%0AOperating system:%0D%0A%0D%0AProgramming language:%0D%0A%0D%0AFeedback:">support@sourcegraph.com</a></strong> with ideas on how we can make Sourcegraph Desktop better.
					</p>
					<p>Did you know we use Slack for feedback and bugs? Let us know if you'd like to join our Slack channel!</p>
				</div>
			</div>
		);
	}
}
