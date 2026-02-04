import * as pulumi from "@pulumi/pulumi";
import * as inputs from "../types/input";
import * as outputs from "../types/output";
/**
 * Manages a GitHub connector in Dex. This connector allows users to authenticate using their GitHub accounts and supports organization and team-based access control.
 */
export declare class GitHubConnector extends pulumi.CustomResource {
    /**
     * Get an existing GitHubConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): GitHubConnector;
    /**
     * Returns true if the given object is an instance of GitHubConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is GitHubConnector;
    /**
     * GitHub OAuth app client ID.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * GitHub OAuth app client secret.
     */
    readonly clientSecret: pulumi.Output<string>;
    /**
     * Unique identifier for the GitHub connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * GitHub Enterprise hostname (e.g., 'github.example.com'). Leave empty for github.com.
     */
    readonly hostName: pulumi.Output<string | undefined>;
    /**
     * If true, load all groups (teams) the user is a member of. Defaults to false.
     */
    readonly loadAllGroups: pulumi.Output<boolean | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * List of GitHub organizations with optional team restrictions. Only users in these orgs/teams will be allowed to authenticate.
     */
    readonly orgs: pulumi.Output<outputs.resources.GitHubOrg[] | undefined>;
    /**
     * Preferred email domain. If set, users with emails in this domain will be preferred.
     */
    readonly preferredEmailDomain: pulumi.Output<string | undefined>;
    /**
     * Redirect URI registered in GitHub OAuth app. Must match Dex's callback URL.
     */
    readonly redirectUri: pulumi.Output<string>;
    /**
     * Root CA certificate for GitHub Enterprise (PEM format). Required if using self-signed certificates.
     */
    readonly rootCA: pulumi.Output<string | undefined>;
    /**
     * Field to use for team names in group claims. Valid values: 'name', 'slug', or 'both'. Defaults to 'slug'.
     */
    readonly teamNameField: pulumi.Output<string | undefined>;
    /**
     * If true, use GitHub login username as the user ID. Defaults to false.
     */
    readonly useLoginAsID: pulumi.Output<boolean | undefined>;
    /**
     * Create a GitHubConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: GitHubConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a GitHubConnector resource.
 */
export interface GitHubConnectorArgs {
    /**
     * GitHub OAuth app client ID.
     */
    clientId: pulumi.Input<string>;
    /**
     * GitHub OAuth app client secret.
     */
    clientSecret: pulumi.Input<string>;
    /**
     * Unique identifier for the GitHub connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * GitHub Enterprise hostname (e.g., 'github.example.com'). Leave empty for github.com.
     */
    hostName?: pulumi.Input<string>;
    /**
     * If true, load all groups (teams) the user is a member of. Defaults to false.
     */
    loadAllGroups?: pulumi.Input<boolean>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * List of GitHub organizations with optional team restrictions. Only users in these orgs/teams will be allowed to authenticate.
     */
    orgs?: pulumi.Input<pulumi.Input<inputs.resources.GitHubOrgArgs>[]>;
    /**
     * Preferred email domain. If set, users with emails in this domain will be preferred.
     */
    preferredEmailDomain?: pulumi.Input<string>;
    /**
     * Redirect URI registered in GitHub OAuth app. Must match Dex's callback URL.
     */
    redirectUri: pulumi.Input<string>;
    /**
     * Root CA certificate for GitHub Enterprise (PEM format). Required if using self-signed certificates.
     */
    rootCA?: pulumi.Input<string>;
    /**
     * Field to use for team names in group claims. Valid values: 'name', 'slug', or 'both'. Defaults to 'slug'.
     */
    teamNameField?: pulumi.Input<string>;
    /**
     * If true, use GitHub login username as the user ID. Defaults to false.
     */
    useLoginAsID?: pulumi.Input<boolean>;
}
