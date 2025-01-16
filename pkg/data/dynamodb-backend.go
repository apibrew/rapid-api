package data

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"regexp"
	"strconv"
	"strings"
)

const recordPK = "record"

var isNumber = regexp.MustCompile(`^\d+$`)

type dynamodbInterface struct {
	cfg DynamodbConfig
	svc *dynamodb.Client
}

func (d *dynamodbInterface) DeleteRecord(path string) error {
	var batchSize = 100

	for {
		output, err := d.svc.Query(context.TODO(), &dynamodb.QueryInput{
			TableName:              aws.String(d.cfg.TableName),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: recordPK},
				":sk": &types.AttributeValueMemberS{Value: path},
			},
			Limit: aws.Int32(int32(batchSize)),
		})

		if err != nil {
			return fmt.Errorf("failed to query items, %v", err)
		}

		for _, item := range output.Items {
			_, err := d.svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
				TableName: aws.String(d.cfg.TableName),
				Key: map[string]types.AttributeValue{
					"PK": item["PK"],
					"SK": item["SK"],
				},
			})

			if err != nil {
				return fmt.Errorf("failed to delete item, %v", err)
			}
		}

		if len(output.Items) < batchSize {
			break
		}
	}

	return nil
}

func (d *dynamodbInterface) WriteRecord(path string, record Record) (Record, error) {
	var processedRecord Record = make(Record)

	for key, value := range record {
		switch value.(type) { // json types
		case map[string]interface{}:
			var subRecord Record = value.(map[string]interface{})
			var subPath = fmt.Sprintf("%s/%s", path, key)

			if subRecord["path"] != nil && strings.HasPrefix(subRecord["path"].(string), subPath) {
				subPath = subRecord["path"].(string)
			}

			_, err := d.WriteRecord(subPath, subRecord)

			if err != nil {
				return Record{}, fmt.Errorf("failed to write sub record, %v", err)
			}
		case []interface{}:
			for _, subRecord := range value.([]interface{}) {
				subRecordMap := subRecord.(map[string]interface{})
				var subPath = fmt.Sprintf("%s/%s", path, key)

				if subRecordMap["path"] != nil && strings.HasPrefix(subRecordMap["path"].(string), subPath) {
					subPath = subRecordMap["path"].(string)
				}

				_, err := d.WriteRecord(subPath, subRecordMap)

				if err != nil {
					return Record{}, fmt.Errorf("failed to write sub record, %v", err)
				}
			}

		default:
			processedRecord[key] = value
		}
	}

	if len(processedRecord) == 0 {
		return nil, nil
	}

	return d.write(path, processedRecord)
}

func (d *dynamodbInterface) write(path string, record Record) (Record, error) {
	isCollection := d.isCollectionPath(path)

	if isCollection {
		lastPath, err := d.getLastRecordPath(path)

		if err != nil {
			return Record{}, fmt.Errorf("failed to get last record path, %v", err)
		}

		lastPathParts := strings.Split(lastPath, "/")
		pathIdPart := lastPathParts[len(lastPathParts)-1]

		if isNumber.MatchString(pathIdPart) {
			var pathId, _ = strconv.Atoi(pathIdPart)
			pathId++
			path = fmt.Sprintf("%s/%d", path, pathId)
		} else {
			path = fmt.Sprintf("%s/%d", path, 1)
		}
	}

	var item, err = attributevalue.MarshalMap(record)

	if err != nil {
		return Record{}, fmt.Errorf("failed to map record to attribute map, %v", err)
	}

	item["PK"] = &types.AttributeValueMemberS{Value: recordPK}
	item["SK"] = &types.AttributeValueMemberS{Value: path}

	_, err = d.svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(d.cfg.TableName),
		Item:      item,
	})

	if err != nil {
		return Record{}, fmt.Errorf("failed to put item, %v", err)
	}

	record["path"] = path

	return record, nil
}

func (d *dynamodbInterface) isCollectionPath(path string) bool {
	var pathParts = strings.Split(path, "/")
	pathParts = pathParts[1:len(pathParts)]
	lastPathPart := pathParts[len(pathParts)-1]

	isCollection := !isNumber.MatchString(lastPathPart)
	return isCollection
}

func (d *dynamodbInterface) getLastRecordPath(path string) (string, error) {
	output, err := d.svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(d.cfg.TableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: recordPK},
			":sk": &types.AttributeValueMemberS{Value: path},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(1),
	})

	if err != nil {
		return "", fmt.Errorf("failed to query items, %v", err)
	}

	if len(output.Items) == 0 {
		return "", nil
	}

	return output.Items[0]["SK"].(*types.AttributeValueMemberS).Value, nil
}

func (d *dynamodbInterface) GetRecords(path string) ([]Record, bool, error) {
	isCollection := d.isCollectionPath(path)

	output, err := d.svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(d.cfg.TableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: recordPK},
			":sk": &types.AttributeValueMemberS{Value: path},
		},
		Limit: aws.Int32(1000),
	})

	if err != nil {
		return nil, isCollection, fmt.Errorf("failed to get item, %v", err)
	}

	if len(output.Items) == 0 {
		return nil, isCollection, nil
	}

	var records []Record

	for _, item := range output.Items {
		record, err := d.mapToRecord(item)

		if err != nil {
			return nil, isCollection, fmt.Errorf("failed to map record, %v", err)
		}

		delete(*record, "PK")
		delete(*record, "SK")

		(*record)["path"] = item["SK"].(*types.AttributeValueMemberS).Value

		records = append(records, *record)
	}

	if len(records) > 1 {
		records, err = d.mergeRecords(path, records)

		if err != nil {
			return nil, isCollection, fmt.Errorf("failed to merge records, %v", err)
		}
	}

	return records, isCollection, nil
}

func (d *dynamodbInterface) mergeRecords(path string, records []Record) ([]Record, error) {
	var mergedRecords []Record

	var recordPathMap = make(map[string]Record)
	var recordChildrenMap = make(map[string][]Record)

	for i := 0; i < len(records); i++ {
		var foundParentPath string
		var record = records[i]
		for parentPath := range recordPathMap {
			if len(parentPath) <= len(foundParentPath) {
				continue
			}

			var isChild = strings.HasPrefix(record["path"].(string), parentPath+"/")

			if isChild {
				// merge record
				foundParentPath = parentPath
				break
			}
		}

		recordPathMap[record["path"].(string)] = record
		if foundParentPath == "" {
			mergedRecords = append(mergedRecords, record)
		} else {
			recordChildrenMap[foundParentPath] = append(recordChildrenMap[foundParentPath], record)
		}
	}

	for parentPath, children := range recordChildrenMap {
		var pathMap = make(map[string][]Record)

		for _, child := range children {
			var rightPath = strings.TrimPrefix(child["path"].(string), parentPath+"/")

			if !d.isCollectionPath(rightPath) {
				rightPath = rightPath[:strings.LastIndex(rightPath, "/")]
			}

			pathMap[rightPath] = append(pathMap[rightPath], child)
		}

		for rightPath, pathRecords := range pathMap {
			if strings.Contains(rightPath, "/") {
				rightPathParts := strings.Split(rightPath, "/")
				var element = recordPathMap[parentPath]

				for i := 0; i < len(rightPathParts)-1; i++ {
					var rightPathPart = rightPathParts[i]

					element[rightPathPart] = make(Record)
					element = element[rightPathPart].(Record)
				}

				element[rightPathParts[len(rightPathParts)-1]] = pathRecords
			} else {
				recordPathMap[parentPath][rightPath] = pathRecords
			}
		}
	}

	return mergedRecords, nil
}

func (d *dynamodbInterface) Init() error {
	var opts []func(*config.LoadOptions) error

	opts = append(opts, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     d.cfg.AccessKeyID,
			SecretAccessKey: d.cfg.SecretAccessKey,
		},
	}))

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return fmt.Errorf("failed to load configuration, %v", err)
	}

	cfg.Region = d.cfg.Region

	d.svc = dynamodb.NewFromConfig(cfg)

	return nil
}

func (d *dynamodbInterface) mapToRecord(data map[string]types.AttributeValue) (*Record, error) {
	var record = new(Record)
	err := attributevalue.UnmarshalMap(data, record)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal record, %v", err)
	}

	return record, nil
}

func NewDynamoDBInterface(cfg DynamodbConfig) Interface {
	return &dynamodbInterface{
		cfg: cfg,
	}
}
