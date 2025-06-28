output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.main.id
}

output "vpc_cidr" {
  description = "The CIDR block of the VPC"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "List of IDs of private subnets"
  value       = aws_subnet.private[*].id
}

output "public_subnet_cidrs" {
  description = "List of CIDR blocks of public subnets"
  value       = aws_subnet.public[*].cidr_block
}

output "private_subnet_cidrs" {
  description = "List of CIDR blocks of private subnets"
  value       = aws_subnet.private[*].cidr_block
}

output "internet_gateway_id" {
  description = "ID of the internet gateway"
  value       = aws_internet_gateway.main.id
}

output "nat_gateway_id" {
  description = "ID of the NAT gateway (if created)"
  value       = var.create_nat_gateway ? aws_nat_gateway.main[0].id : null
}

output "public_route_table_id" {
  description = "ID of the public route table"
  value       = aws_route_table.public.id
}

output "private_route_table_id" {
  description = "ID of the private route table"
  value       = aws_route_table.private.id
}

output "alb_security_group_id" {
  description = "ID of the security group for the ALB"
  value       = aws_security_group.alb.id
}

output "ecs_security_group_id" {
  description = "ID of the legacy security group for ECS tasks (deprecated - use service-specific security groups)"
  value       = aws_security_group.ecs_react.id
}

output "ecs_react_security_group_id" {
  description = "ID of the security group for React ECS tasks"
  value       = aws_security_group.ecs_react.id
}

output "availability_zones" {
  description = "List of availability zones used"
  value       = var.availability_zones
}
