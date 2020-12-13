<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: graphik.proto

namespace Api;

use Google\Protobuf\Internal\GPBType;
use Google\Protobuf\Internal\RepeatedField;
use Google\Protobuf\Internal\GPBUtil;

/**
 * Generated from protobuf message <code>api.Flags</code>
 */
class Flags extends \Google\Protobuf\Internal\Message
{
    /**
     * open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID)
     *
     * Generated from protobuf field <code>string open_id_discovery = 1;</code>
     */
    private $open_id_discovery = '';
    /**
     * persistant storage ref (env: GRAPHIK_STORAGE_PATH)
     *
     * Generated from protobuf field <code>string storage_path = 2;</code>
     */
    private $storage_path = '';
    /**
     * enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)
     *
     * Generated from protobuf field <code>bool metrics = 3;</code>
     */
    private $metrics = false;
    /**
     * cors allow headers (env: GRAPHIK_ALLOW_HEADERS)
     *
     * Generated from protobuf field <code>repeated string allow_headers = 5;</code>
     */
    private $allow_headers;
    /**
     * cors allow methods (env: GRAPHIK_ALLOW_METHODS)
     *
     * Generated from protobuf field <code>repeated string allow_methods = 6;</code>
     */
    private $allow_methods;
    /**
     * cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)
     *
     * Generated from protobuf field <code>repeated string allow_origins = 7;</code>
     */
    private $allow_origins;
    /**
     * root user is a list of email addresses that bypass authorizers. (env: GRAPHIK_ROOT_USERS)
     *
     * Generated from protobuf field <code>repeated string root_users = 8;</code>
     */
    private $root_users;
    /**
     * Generated from protobuf field <code>string tls_cert = 9;</code>
     */
    private $tls_cert = '';
    /**
     * Generated from protobuf field <code>string tls_key = 10;</code>
     */
    private $tls_key = '';
    /**
     * Generated from protobuf field <code>string playground_client_id = 11;</code>
     */
    private $playground_client_id = '';
    /**
     * Generated from protobuf field <code>string playground_client_secret = 12;</code>
     */
    private $playground_client_secret = '';
    /**
     * Generated from protobuf field <code>string playground_redirect = 13;</code>
     */
    private $playground_redirect = '';
    /**
     * Generated from protobuf field <code>string playground_session_store = 14;</code>
     */
    private $playground_session_store = '';

    /**
     * Constructor.
     *
     * @param array $data {
     *     Optional. Data for populating the Message object.
     *
     *     @type string $open_id_discovery
     *           open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID)
     *     @type string $storage_path
     *           persistant storage ref (env: GRAPHIK_STORAGE_PATH)
     *     @type bool $metrics
     *           enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)
     *     @type string[]|\Google\Protobuf\Internal\RepeatedField $allow_headers
     *           cors allow headers (env: GRAPHIK_ALLOW_HEADERS)
     *     @type string[]|\Google\Protobuf\Internal\RepeatedField $allow_methods
     *           cors allow methods (env: GRAPHIK_ALLOW_METHODS)
     *     @type string[]|\Google\Protobuf\Internal\RepeatedField $allow_origins
     *           cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)
     *     @type string[]|\Google\Protobuf\Internal\RepeatedField $root_users
     *           root user is a list of email addresses that bypass authorizers. (env: GRAPHIK_ROOT_USERS)
     *     @type string $tls_cert
     *     @type string $tls_key
     *     @type string $playground_client_id
     *     @type string $playground_client_secret
     *     @type string $playground_redirect
     *     @type string $playground_session_store
     * }
     */
    public function __construct($data = NULL) {
        \GPBMetadata\Graphik::initOnce();
        parent::__construct($data);
    }

    /**
     * open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID)
     *
     * Generated from protobuf field <code>string open_id_discovery = 1;</code>
     * @return string
     */
    public function getOpenIdDiscovery()
    {
        return $this->open_id_discovery;
    }

    /**
     * open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID)
     *
     * Generated from protobuf field <code>string open_id_discovery = 1;</code>
     * @param string $var
     * @return $this
     */
    public function setOpenIdDiscovery($var)
    {
        GPBUtil::checkString($var, True);
        $this->open_id_discovery = $var;

        return $this;
    }

    /**
     * persistant storage ref (env: GRAPHIK_STORAGE_PATH)
     *
     * Generated from protobuf field <code>string storage_path = 2;</code>
     * @return string
     */
    public function getStoragePath()
    {
        return $this->storage_path;
    }

    /**
     * persistant storage ref (env: GRAPHIK_STORAGE_PATH)
     *
     * Generated from protobuf field <code>string storage_path = 2;</code>
     * @param string $var
     * @return $this
     */
    public function setStoragePath($var)
    {
        GPBUtil::checkString($var, True);
        $this->storage_path = $var;

        return $this;
    }

    /**
     * enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)
     *
     * Generated from protobuf field <code>bool metrics = 3;</code>
     * @return bool
     */
    public function getMetrics()
    {
        return $this->metrics;
    }

    /**
     * enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)
     *
     * Generated from protobuf field <code>bool metrics = 3;</code>
     * @param bool $var
     * @return $this
     */
    public function setMetrics($var)
    {
        GPBUtil::checkBool($var);
        $this->metrics = $var;

        return $this;
    }

    /**
     * cors allow headers (env: GRAPHIK_ALLOW_HEADERS)
     *
     * Generated from protobuf field <code>repeated string allow_headers = 5;</code>
     * @return \Google\Protobuf\Internal\RepeatedField
     */
    public function getAllowHeaders()
    {
        return $this->allow_headers;
    }

    /**
     * cors allow headers (env: GRAPHIK_ALLOW_HEADERS)
     *
     * Generated from protobuf field <code>repeated string allow_headers = 5;</code>
     * @param string[]|\Google\Protobuf\Internal\RepeatedField $var
     * @return $this
     */
    public function setAllowHeaders($var)
    {
        $arr = GPBUtil::checkRepeatedField($var, \Google\Protobuf\Internal\GPBType::STRING);
        $this->allow_headers = $arr;

        return $this;
    }

    /**
     * cors allow methods (env: GRAPHIK_ALLOW_METHODS)
     *
     * Generated from protobuf field <code>repeated string allow_methods = 6;</code>
     * @return \Google\Protobuf\Internal\RepeatedField
     */
    public function getAllowMethods()
    {
        return $this->allow_methods;
    }

    /**
     * cors allow methods (env: GRAPHIK_ALLOW_METHODS)
     *
     * Generated from protobuf field <code>repeated string allow_methods = 6;</code>
     * @param string[]|\Google\Protobuf\Internal\RepeatedField $var
     * @return $this
     */
    public function setAllowMethods($var)
    {
        $arr = GPBUtil::checkRepeatedField($var, \Google\Protobuf\Internal\GPBType::STRING);
        $this->allow_methods = $arr;

        return $this;
    }

    /**
     * cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)
     *
     * Generated from protobuf field <code>repeated string allow_origins = 7;</code>
     * @return \Google\Protobuf\Internal\RepeatedField
     */
    public function getAllowOrigins()
    {
        return $this->allow_origins;
    }

    /**
     * cors allow origins (env: GRAPHIK_ALLOW_ORIGINS)
     *
     * Generated from protobuf field <code>repeated string allow_origins = 7;</code>
     * @param string[]|\Google\Protobuf\Internal\RepeatedField $var
     * @return $this
     */
    public function setAllowOrigins($var)
    {
        $arr = GPBUtil::checkRepeatedField($var, \Google\Protobuf\Internal\GPBType::STRING);
        $this->allow_origins = $arr;

        return $this;
    }

    /**
     * root user is a list of email addresses that bypass authorizers. (env: GRAPHIK_ROOT_USERS)
     *
     * Generated from protobuf field <code>repeated string root_users = 8;</code>
     * @return \Google\Protobuf\Internal\RepeatedField
     */
    public function getRootUsers()
    {
        return $this->root_users;
    }

    /**
     * root user is a list of email addresses that bypass authorizers. (env: GRAPHIK_ROOT_USERS)
     *
     * Generated from protobuf field <code>repeated string root_users = 8;</code>
     * @param string[]|\Google\Protobuf\Internal\RepeatedField $var
     * @return $this
     */
    public function setRootUsers($var)
    {
        $arr = GPBUtil::checkRepeatedField($var, \Google\Protobuf\Internal\GPBType::STRING);
        $this->root_users = $arr;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string tls_cert = 9;</code>
     * @return string
     */
    public function getTlsCert()
    {
        return $this->tls_cert;
    }

    /**
     * Generated from protobuf field <code>string tls_cert = 9;</code>
     * @param string $var
     * @return $this
     */
    public function setTlsCert($var)
    {
        GPBUtil::checkString($var, True);
        $this->tls_cert = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string tls_key = 10;</code>
     * @return string
     */
    public function getTlsKey()
    {
        return $this->tls_key;
    }

    /**
     * Generated from protobuf field <code>string tls_key = 10;</code>
     * @param string $var
     * @return $this
     */
    public function setTlsKey($var)
    {
        GPBUtil::checkString($var, True);
        $this->tls_key = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string playground_client_id = 11;</code>
     * @return string
     */
    public function getPlaygroundClientId()
    {
        return $this->playground_client_id;
    }

    /**
     * Generated from protobuf field <code>string playground_client_id = 11;</code>
     * @param string $var
     * @return $this
     */
    public function setPlaygroundClientId($var)
    {
        GPBUtil::checkString($var, True);
        $this->playground_client_id = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string playground_client_secret = 12;</code>
     * @return string
     */
    public function getPlaygroundClientSecret()
    {
        return $this->playground_client_secret;
    }

    /**
     * Generated from protobuf field <code>string playground_client_secret = 12;</code>
     * @param string $var
     * @return $this
     */
    public function setPlaygroundClientSecret($var)
    {
        GPBUtil::checkString($var, True);
        $this->playground_client_secret = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string playground_redirect = 13;</code>
     * @return string
     */
    public function getPlaygroundRedirect()
    {
        return $this->playground_redirect;
    }

    /**
     * Generated from protobuf field <code>string playground_redirect = 13;</code>
     * @param string $var
     * @return $this
     */
    public function setPlaygroundRedirect($var)
    {
        GPBUtil::checkString($var, True);
        $this->playground_redirect = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string playground_session_store = 14;</code>
     * @return string
     */
    public function getPlaygroundSessionStore()
    {
        return $this->playground_session_store;
    }

    /**
     * Generated from protobuf field <code>string playground_session_store = 14;</code>
     * @param string $var
     * @return $this
     */
    public function setPlaygroundSessionStore($var)
    {
        GPBUtil::checkString($var, True);
        $this->playground_session_store = $var;

        return $this;
    }

}

