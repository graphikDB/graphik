<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: graphik.proto

namespace Api;

use Google\Protobuf\Internal\GPBType;
use Google\Protobuf\Internal\RepeatedField;
use Google\Protobuf\Internal\GPBUtil;

/**
 * Docs is an array of docs
 *
 * Generated from protobuf message <code>api.Docs</code>
 */
class Docs extends \Google\Protobuf\Internal\Message
{
    /**
     * docs is an array of docs
     *
     * Generated from protobuf field <code>repeated .api.Doc docs = 1;</code>
     */
    private $docs;
    /**
     * Generated from protobuf field <code>string seek_next = 2;</code>
     */
    private $seek_next = '';

    /**
     * Constructor.
     *
     * @param array $data {
     *     Optional. Data for populating the Message object.
     *
     *     @type \Api\Doc[]|\Google\Protobuf\Internal\RepeatedField $docs
     *           docs is an array of docs
     *     @type string $seek_next
     * }
     */
    public function __construct($data = NULL) {
        \GPBMetadata\Graphik::initOnce();
        parent::__construct($data);
    }

    /**
     * docs is an array of docs
     *
     * Generated from protobuf field <code>repeated .api.Doc docs = 1;</code>
     * @return \Google\Protobuf\Internal\RepeatedField
     */
    public function getDocs()
    {
        return $this->docs;
    }

    /**
     * docs is an array of docs
     *
     * Generated from protobuf field <code>repeated .api.Doc docs = 1;</code>
     * @param \Api\Doc[]|\Google\Protobuf\Internal\RepeatedField $var
     * @return $this
     */
    public function setDocs($var)
    {
        $arr = GPBUtil::checkRepeatedField($var, \Google\Protobuf\Internal\GPBType::MESSAGE, \Api\Doc::class);
        $this->docs = $arr;

        return $this;
    }

    /**
     * Generated from protobuf field <code>string seek_next = 2;</code>
     * @return string
     */
    public function getSeekNext()
    {
        return $this->seek_next;
    }

    /**
     * Generated from protobuf field <code>string seek_next = 2;</code>
     * @param string $var
     * @return $this
     */
    public function setSeekNext($var)
    {
        GPBUtil::checkString($var, True);
        $this->seek_next = $var;

        return $this;
    }

}
