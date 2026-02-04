import * as pulumi from "@pulumi/pulumi";
/**
 * Manages a GitLab connector in Dex. This connector allows users to authenticate using their GitLab accounts and supports group-based access control.
 */
export declare class GitLabConnector extends pulumi.CustomResource {
    /**
     * Get an existing GitLabConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): GitLabConnector;
    /**
     * Returns true if the given object is an instance of GitLabConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is GitLabConnector;
    /**
     * GitLab instance base URL. Defaults to 'https://gitlab.com' for GitLab.com.
     */
    readonly baseURL: pulumi.Output<string | undefined>;
    /**
     * GitLab OAuth application client ID.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * GitLab OAuth application client secret.
     */
    readonly clientSecret: pulumi.Output<string>;
    /**
     * Unique identifier for the GitLab connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * If true, request 'read_api' scope to fetch group memberships. Defaults to false.
     */
    readonly getGroupsPermission: pulumi.Output<boolean | undefined>;
    /**
     * List of GitLab group names. Only users in these groups will be allowed to authenticate.
     */
    readonly groups: pulumi.Output<string[] | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * Redirect URI registered in GitLab OAuth app. Must match Dex's callback URL.
     */
    readonly redirectUri: pulumi.Output<string>;
    /**
     * If true, use GitLab username as the user ID. Defaults to false.
     */
    readonly useLoginAsID: pulumi.Output<boolean | undefined>;
    /**
     * Create a GitLabConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: GitLabConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a GitLabConnector resource.
 */
export interface GitLabConnectorArgs {
    /**
     * GitLab instance base URL. Defaults to 'https://gitlab.com' for GitLab.com.
     */
    baseURL?: pulumi.Input<string>;
    /**
     * GitLab OAuth application client ID.
     */
    clientId: pulumi.Input<string>;
    /**
     * GitLab OAuth application client secret.
     */
    clientSecret: pulumi.Input<string>;
    /**
     * Unique identifier for the GitLab connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * If true, request 'read_api' scope to fetch group memberships. Defaults to false.
     */
    getGroupsPermission?: pulumi.Input<boolean>;
    /**
     * List of GitLab group names. Only users in these groups will be allowed to authenticate.
     */
    groups?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * Redirect URI registered in GitLab OAuth app. Must match Dex's callback URL.
     */
    redirectUri: pulumi.Input<string>;
    /**
     * If true, use GitLab username as the user ID. Defaults to false.
     */
    useLoginAsID?: pulumi.Input<boolean>;
}
