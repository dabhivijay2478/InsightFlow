# AWS GitHub OIDC Role Guide

Use GitHub OIDC for every production workflow. Do not put long-lived AWS access keys in GitHub secrets.

This guide creates three roles:

- `mantrixflow-infra-github-actions`
- `mantrixflow-api-github-actions`
- `mantrixflow-elt-github-actions`

Save the role ARNs into GitHub environment secrets:

- Infra repo `production-infra`: `AWS_INFRA_DEPLOY_ROLE_ARN`
- API repo `production-api`: `AWS_API_DEPLOY_ROLE_ARN`
- ELT repo `production-elt`: `AWS_ELT_DEPLOY_ROLE_ARN`

## OIDC Provider

Create this provider once per AWS account:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com
```

If the CLI requires a thumbprint:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

## Infra Role Trust Policy

Use this trust policy for `mantrixflow-infra-github-actions`.

Because the workflow uses the GitHub environment `production-infra`, the OIDC `sub` is environment-scoped. Restrict the environment itself to the `main` branch in GitHub.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub": "repo:dabhivijay2478/mantrixflow-infra:environment:production-infra"
        }
      }
    }
  ]
}
```

## API Role Trust Policy

Use this trust policy for `mantrixflow-api-github-actions`.

Restrict the GitHub environment `production-api` to the `production` branch.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub": "repo:dabhivijay2478/cloud.api.mantrixflow.com:environment:production-api"
        }
      }
    }
  ]
}
```

## ELT Role Trust Policy

Use this trust policy for `mantrixflow-elt-github-actions`.

Restrict the GitHub environment `production-elt` to the `production` branch.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub": "repo:dabhivijay2478/cloud.api.etl.server.mantrixflow.com:environment:production-elt"
        }
      }
    }
  ]
}
```

## Infra Starter Permission Policy

This role runs CDK and Terraform for shared production resources. Use this broad starter policy for the first launch, then tighten it after the first successful production deploy.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CdkAndTerraformProductionResources",
      "Effect": "Allow",
      "Action": [
        "acm:*",
        "application-autoscaling:*",
        "cloudformation:*",
        "ec2:*",
        "ecr:*",
        "ecs:*",
        "elasticloadbalancing:*",
        "iam:GetRole",
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:PassRole",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:PutRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:GetRolePolicy",
        "iam:ListRolePolicies",
        "iam:ListAttachedRolePolicies",
        "logs:*",
        "route53:ListHostedZones",
        "servicediscovery:*",
        "ssm:*",
        "sts:GetCallerIdentity"
      ],
      "Resource": "*"
    },
    {
      "Sid": "TerraformStateBucket",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::mantrixflow-tfstate",
        "arn:aws:s3:::mantrixflow-tfstate/*"
      ]
    }
  ]
}
```

## API Deploy Permission Policy

This role only pushes the API image and updates `mantrixflow-api-service`.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EcrLogin",
      "Effect": "Allow",
      "Action": "ecr:GetAuthorizationToken",
      "Resource": "*"
    },
    {
      "Sid": "PushApiImage",
      "Effect": "Allow",
      "Action": [
        "ecr:BatchCheckLayerAvailability",
        "ecr:CompleteLayerUpload",
        "ecr:DescribeImages",
        "ecr:DescribeRepositories",
        "ecr:InitiateLayerUpload",
        "ecr:PutImage",
        "ecr:UploadLayerPart"
      ],
      "Resource": "arn:aws:ecr:ap-south-1:AWS_ACCOUNT_ID:repository/mantrixflow-api"
    },
    {
      "Sid": "UpdateApiService",
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:RegisterTaskDefinition",
        "ecs:UpdateService"
      ],
      "Resource": "*"
    },
    {
      "Sid": "PassExistingEcsRoles",
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:aws:iam::AWS_ACCOUNT_ID:role/MantrixflowEcs-*ExecutionRole*",
      "Condition": {
        "StringEquals": {
          "iam:PassedToService": "ecs-tasks.amazonaws.com"
        }
      }
    }
  ]
}
```

## ELT Deploy Permission Policy

This role only pushes the ELT image and updates `mantrixflow-elt-service`.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EcrLogin",
      "Effect": "Allow",
      "Action": "ecr:GetAuthorizationToken",
      "Resource": "*"
    },
    {
      "Sid": "PushEltImage",
      "Effect": "Allow",
      "Action": [
        "ecr:BatchCheckLayerAvailability",
        "ecr:CompleteLayerUpload",
        "ecr:DescribeImages",
        "ecr:DescribeRepositories",
        "ecr:InitiateLayerUpload",
        "ecr:PutImage",
        "ecr:UploadLayerPart"
      ],
      "Resource": "arn:aws:ecr:ap-south-1:AWS_ACCOUNT_ID:repository/mantrixflow-elt"
    },
    {
      "Sid": "UpdateEltService",
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:RegisterTaskDefinition",
        "ecs:UpdateService"
      ],
      "Resource": "*"
    },
    {
      "Sid": "PassExistingEcsRoles",
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:aws:iam::AWS_ACCOUNT_ID:role/MantrixflowEcs-*ExecutionRole*",
      "Condition": {
        "StringEquals": {
          "iam:PassedToService": "ecs-tasks.amazonaws.com"
        }
      }
    }
  ]
}
```

## Branch-Based Trust Alternative

If a workflow does not use a GitHub environment, use branch-scoped `sub` values instead:

- Infra: `repo:dabhivijay2478/mantrixflow-infra:ref:refs/heads/main`
- API: `repo:dabhivijay2478/cloud.api.mantrixflow.com:ref:refs/heads/production`
- ELT: `repo:dabhivijay2478/cloud.api.etl.server.mantrixflow.com:ref:refs/heads/production`

The current workflows use environments, so the environment-scoped trust policies above are the correct default.
