import * as pulumi from "@pulumi/pulumi";
/**
 * Manages an AWS Cognito user pool connector in Dex using the generic OIDC connector (type: oidc). This connector allows users to authenticate using their AWS Cognito credentials.
 */
export declare class CognitoOidcConnector extends pulumi.CustomResource {
    /**
     * Get an existing CognitoOidcConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): CognitoOidcConnector;
    /**
     * Returns true if the given object is an instance of CognitoOidcConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is CognitoOidcConnector;
    /**
     * Cognito app client ID.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * Cognito app client secret.
     */
    readonly clientSecret: pulumi.Output<string>;
    /**
     * Unique identifier for the Cognito connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * Additional OIDC configuration fields as key-value pairs for advanced scenarios.
     */
    readonly extraOidc: pulumi.Output<{
        [key: string]: any;
    } | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * Redirect URI registered in Cognito. Must match Dex's callback URL.
     */
    readonly redirectUri: pulumi.Output<string>;
    /**
     * AWS region where the Cognito user pool is located (e.g., 'us-east-1', 'eu-west-1').
     */
    readonly region: pulumi.Output<string>;
    /**
     * OIDC scopes to request from Cognito. Defaults to ['openid', 'email', 'profile'] if not specified.
     */
    readonly scopes: pulumi.Output<string[] | undefined>;
    /**
     * Source for the username claim. Valid values: 'email' or 'sub' (subject).
     */
    readonly userNameSource: pulumi.Output<string | undefined>;
    /**
     * AWS Cognito user pool ID.
     */
    readonly userPoolId: pulumi.Output<string>;
    /**
     * Create a CognitoOidcConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: CognitoOidcConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a CognitoOidcConnector resource.
 */
export interface CognitoOidcConnectorArgs {
    /**
     * Cognito app client ID.
     */
    clientId: pulumi.Input<string>;
    /**
     * Cognito app client secret.
     */
    clientSecret: pulumi.Input<string>;
    /**
     * Unique identifier for the Cognito connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * Additional OIDC configuration fields as key-value pairs for advanced scenarios.
     */
    extraOidc?: pulumi.Input<{
        [key: string]: any;
    }>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * Redirect URI registered in Cognito. Must match Dex's callback URL.
     */
    redirectUri: pulumi.Input<string>;
    /**
     * AWS region where the Cognito user pool is located (e.g., 'us-east-1', 'eu-west-1').
     */
    region: pulumi.Input<string>;
    /**
     * OIDC scopes to request from Cognito. Defaults to ['openid', 'email', 'profile'] if not specified.
     */
    scopes?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * Source for the username claim. Valid values: 'email' or 'sub' (subject).
     */
    userNameSource?: pulumi.Input<string>;
    /**
     * AWS Cognito user pool ID.
     */
    userPoolId: pulumi.Input<string>;
}
