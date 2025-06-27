# Application Load Balancer for React

# ALB
resource "aws_lb" "react" {
  name               = "vm-hub-${local.name_suffix}-alb"
  internal           = var.internal_alb
  load_balancer_type = "application"
  security_groups    = [var.alb_security_group_id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = var.enable_deletion_protection
  enable_http2               = true
  idle_timeout               = var.alb_idle_timeout

  dynamic "access_logs" {
    for_each = var.alb_access_logs_bucket != "" ? [1] : []
    content {
      bucket  = var.alb_access_logs_bucket
      prefix  = var.alb_access_logs_prefix != "" ? var.alb_access_logs_prefix : null
      enabled = false
    }
  }

  tags = merge(
    var.common_tags,
    {
      Name = "vm-hub-${local.name_suffix}-alb"
    }
  )
}

# Target Group
resource "aws_lb_target_group" "react" {
  name                 = "${local.short_service_name}-tg"
  port                 = var.port
  protocol             = "HTTP"
  vpc_id               = var.vpc_id
  target_type          = "ip"
  deregistration_delay = var.deregistration_delay

  health_check {
    enabled             = true
    interval            = var.health_check_interval
    path                = var.health_check_path
    port                = "traffic-port"
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold
    timeout             = var.health_check_timeout
    matcher             = "200-399"
  }

  tags = merge(
    var.common_tags,
    {
      Name = "${local.service_name}-tg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# HTTP Listener
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.react.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = var.enable_https ? "redirect" : "forward"

    dynamic "redirect" {
      for_each = var.enable_https ? [1] : []
      content {
        port        = "443"
        protocol    = "HTTPS"
        status_code = "HTTP_301"
      }
    }

    dynamic "forward" {
      for_each = var.enable_https ? [] : [1]
      content {
        target_group {
          arn = aws_lb_target_group.react.arn
        }
      }
    }
  }

  tags = merge(
    var.common_tags,
    {
      Name = "${local.service_name}-http-listener"
    }
  )
}

# HTTPS Listener (if enabled)
resource "aws_lb_listener" "https" {
  count             = var.enable_https ? 1 : 0
  load_balancer_arn = aws_lb.react.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = var.ssl_policy
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.react.arn
  }

  tags = merge(
    var.common_tags,
    {
      Name = "${local.service_name}-https-listener"
    }
  )
}
